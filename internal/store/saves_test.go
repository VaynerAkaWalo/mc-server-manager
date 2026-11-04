package store

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newSave(t *testing.T, s *Store, owner string) SaveRecord {
	t.Helper()

	record, err := s.CreateSave(context.Background(), owner, "world", 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = s.pool.Exec(context.Background(), `DELETE FROM mc_servers WHERE save_id = $1`, record.ID)
		_, _ = s.pool.Exec(context.Background(), `DELETE FROM saves WHERE id = $1`, record.ID)
	})

	return record
}

func serverOnSave(t *testing.T, s *Store, name string, save SaveRecord) (ServerRecord, error) {
	t.Helper()

	record, err := s.CreateServer(context.Background(), NewServer{
		Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour),
		Save: &SaveUse{ID: save.ID, Owner: save.Owner},
	})
	if err == nil {
		t.Cleanup(func() { _ = s.DeleteServer(context.Background(), record.ID) })
	}

	return record, err
}

func TestOneServerHoldsASaveUntilItsLockIsReleased(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "lock-owner")

	first, err := serverOnSave(t, s, "lock-first", save)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverOnSave(t, s, "lock-second", save); !errors.Is(err, ErrSaveInUse) {
		t.Fatalf("second server on a held save: %v", err)
	}

	if err := s.MarkStopped(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := serverOnSave(t, s, "lock-third", save); !errors.Is(err, ErrSaveInUse) {
		t.Fatalf("a stopped server must keep the save until the cluster resource is gone: %v", err)
	}

	releasable, err := s.ListReleasableSaveLocks(ctx)
	if err != nil || len(releasable) != 1 || releasable[0].ID != first.ID || releasable[0].Name != "lock-first" {
		t.Fatalf("releasable = %+v, %v", releasable, err)
	}
	if err := s.ReleaseSaveLock(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := serverOnSave(t, s, "lock-fourth", save); err != nil {
		t.Fatalf("server after release: %v", err)
	}
}

func TestActiveServerLockCannotBeReleased(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "active-owner")
	running, err := serverOnSave(t, s, "active-running", save)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.ReleaseSaveLock(ctx, running.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := serverOnSave(t, s, "active-other", save); !errors.Is(err, ErrSaveInUse) {
		t.Fatalf("a running server lost its save: %v", err)
	}
}

func TestServerNameConflictIsNotReportedAsSaveInUse(t *testing.T) {
	s := connectTestStore(t)
	first := newSave(t, s, "name-owner")
	second := newSave(t, s, "name-owner")
	if _, err := serverOnSave(t, s, "name-clash", first); err != nil {
		t.Fatal(err)
	}

	if _, err := serverOnSave(t, s, "name-clash", second); !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestServerCanOnlyUseASaveItsOwnerOwns(t *testing.T) {
	s := connectTestStore(t)
	save := newSave(t, s, "real-owner")
	stolen := save
	stolen.Owner = "someone-else"

	if _, err := serverOnSave(t, s, "stolen-save", stolen); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := serverOnSave(t, s, "missing-save", SaveRecord{ID: uuid.New(), Owner: "real-owner"}); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestServerWithoutSaveHoldsNothing(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	record, err := s.CreateServer(ctx, NewServer{Name: "no-save", Hostname: "no-save", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })

	var locked bool
	if err := s.pool.QueryRow(ctx, `SELECT save_locked FROM mc_servers WHERE id = $1`, record.ID).Scan(&locked); err != nil || locked {
		t.Fatalf("locked = %v, err = %v", locked, err)
	}
}

func TestSaveLimitAppliesPerOwner(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	owner := "limit-" + uuid.NewString()
	t.Cleanup(func() { _, _ = s.pool.Exec(ctx, `DELETE FROM saves WHERE owner_subject = $1`, owner) })

	for i := 0; i < 2; i++ {
		if _, err := s.CreateSave(ctx, owner, "world", 2); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateSave(ctx, owner, "world", 2); !errors.Is(err, ErrSaveLimit) {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.CreateSave(ctx, owner+"-other", "world", 2); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = s.pool.Exec(ctx, `DELETE FROM saves WHERE owner_subject = $1`, owner+"-other") })
}

func TestConcurrentSaveCreationStaysWithinTheLimit(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	owner := "race-" + uuid.NewString()
	t.Cleanup(func() { _, _ = s.pool.Exec(ctx, `DELETE FROM saves WHERE owner_subject = $1`, owner) })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.CreateSave(ctx, owner, "world", 3)
		}()
	}
	wg.Wait()

	saves, err := s.ListSaves(ctx, owner)
	if err != nil || len(saves) != 3 {
		t.Fatalf("saves = %d, err = %v", len(saves), err)
	}
}

func TestRenameAndGetSaveAreScopedToTheOwner(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "rename-owner")

	if _, err := s.RenameSave(ctx, save.ID, "intruder", "mine"); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("rename by another owner: %v", err)
	}
	renamed, err := s.RenameSave(ctx, save.ID, "rename-owner", "Creative")
	if err != nil || renamed.Name != "Creative" {
		t.Fatalf("renamed = %+v, %v", renamed, err)
	}
	if _, err := s.GetSave(ctx, save.ID, "intruder"); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("get by another owner: %v", err)
	}
}

func TestDeleteIdleSaveRechecksActivityBeforePurge(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "idle-owner")

	purged := false
	purge := func(context.Context) error { purged = true; return nil }
	if err := s.DeleteIdleSave(ctx, save.ID, 90*24*time.Hour, purge); !errors.Is(err, ErrSaveInUse) || purged {
		t.Fatalf("fresh save: %v, purged = %v", err, purged)
	}

	if _, err := s.pool.Exec(ctx, `UPDATE saves SET created_at = now() - interval '91 days' WHERE id = $1`, save.ID); err != nil {
		t.Fatal(err)
	}
	running, err := serverOnSave(t, s, "idle-running", save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkActive(ctx, running.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, running.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, running.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteIdleSave(ctx, save.ID, 90*24*time.Hour, purge); !errors.Is(err, ErrSaveInUse) || purged {
		t.Fatalf("recently active save: %v, purged = %v", err, purged)
	}

	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = now() - interval '91 days' WHERE id = $1`, running.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteIdleSave(ctx, save.ID, 90*24*time.Hour, purge); err != nil || !purged {
		t.Fatalf("idle save: %v, purged = %v", err, purged)
	}
}

func TestDeleteSavePurgesBeforeRemovingTheRow(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "delete-owner")

	failing := func(context.Context) error { return errors.New("bucket down") }
	if err := s.DeleteSave(ctx, save.ID, "delete-owner", failing); err == nil {
		t.Fatal("expected the purge failure")
	}
	if _, err := s.GetSave(ctx, save.ID, "delete-owner"); err != nil {
		t.Fatalf("the save row must stay when its data could not be purged: %v", err)
	}

	purged := false
	if err := s.DeleteSave(ctx, save.ID, "delete-owner", func(context.Context) error { purged = true; return nil }); err != nil || !purged {
		t.Fatalf("delete: %v, purged = %v", err, purged)
	}
	if err := s.DeleteSave(ctx, save.ID, "delete-owner", func(context.Context) error { return nil }); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestDeleteSaveRefusesWhileAServerHoldsIt(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "busy-owner")
	if _, err := serverOnSave(t, s, "busy-server", save); err != nil {
		t.Fatal(err)
	}

	purged := false
	err := s.DeleteSave(ctx, save.ID, "busy-owner", func(context.Context) error { purged = true; return nil })

	if !errors.Is(err, ErrSaveInUse) || purged {
		t.Fatalf("err = %v, purged = %v", err, purged)
	}
	if err := s.DeleteSave(ctx, save.ID, "intruder", func(context.Context) error { return nil }); !errors.Is(err, ErrSaveInUse) && !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestDeleteSaveByAnotherOwnerPurgesNothing(t *testing.T) {
	s := connectTestStore(t)
	save := newSave(t, s, "guarded-owner")

	purged := false
	err := s.DeleteSave(context.Background(), save.ID, "intruder", func(context.Context) error { purged = true; return nil })

	if !errors.Is(err, ErrSaveNotFound) || purged {
		t.Fatalf("err = %v, purged = %v", err, purged)
	}
}

func TestConcurrentServerCreationAndSaveDeletionNeverOrphansAServer(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	for attempt := 0; attempt < 10; attempt++ {
		save := newSave(t, s, "race-owner")
		name := "race-server-" + uuid.NewString()[:8]

		var wg sync.WaitGroup
		var createErr, deleteErr error
		var created ServerRecord
		wg.Add(2)
		go func() {
			defer wg.Done()
			created, createErr = s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour), Save: &SaveUse{ID: save.ID, Owner: save.Owner}})
		}()
		go func() {
			defer wg.Done()
			deleteErr = s.DeleteSave(ctx, save.ID, save.Owner, func(context.Context) error { return nil })
		}()
		wg.Wait()

		if createErr == nil {
			_ = s.DeleteServer(ctx, created.ID)
		}
		if createErr == nil && deleteErr == nil {
			t.Fatalf("attempt %d: a server started on a save that was deleted", attempt)
		}
	}
}

func TestLatestSaveDefinitionIgnoresServersThatNeverStarted(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, "definition-owner")

	played := func(name, version string, active bool) {
		t.Helper()
		record, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{"version": version}, ExpiresAt: time.Now().Add(time.Hour), Save: &SaveUse{ID: save.ID, Owner: save.Owner}})
		if err != nil {
			t.Fatal(err)
		}
		if active {
			_, err = s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = now() WHERE id = $1`, record.ID)
		}
		if err == nil {
			err = s.MarkStopped(ctx, record.ID)
		}
		if err == nil {
			err = s.ReleaseSaveLock(ctx, record.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	if _, found, err := s.LatestSaveDefinition(ctx, save.ID); err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}

	played("definition-old", "1.21.3", true)
	played("definition-never-started", "1.21.10", false)

	latest, found, err := s.LatestSaveDefinition(ctx, save.ID)
	if err != nil || !found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
	var definition struct{ Version string }
	if err := json.Unmarshal(latest, &definition); err != nil || definition.Version != "1.21.3" {
		t.Fatalf("latest = %s, %v", latest, err)
	}
}

func TestSavesIdleForTheRetentionPeriodAreListed(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	idle := newSave(t, s, "idle-owner")
	recent := newSave(t, s, "idle-owner")
	held := newSave(t, s, "idle-owner")
	fresh := newSave(t, s, "idle-owner")

	old := time.Now().Add(-100 * 24 * time.Hour)
	for _, save := range []SaveRecord{idle, recent, held} {
		if _, err := s.pool.Exec(ctx, `UPDATE saves SET created_at = $2 WHERE id = $1`, save.ID, old); err != nil {
			t.Fatal(err)
		}
	}

	oldServer, err := serverOnSave(t, s, "idle-played", recent)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = now() - interval '10 days', save_locked = false WHERE id = $1`, oldServer.ID); err != nil {
		t.Fatal(err)
	}
	heldServer, err := serverOnSave(t, s, "idle-held", held)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = now() - interval '200 days' WHERE id = $1`, heldServer.ID); err != nil {
		t.Fatal(err)
	}

	ids, err := s.ListSavesIdleSince(ctx, 90*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	listed := map[uuid.UUID]bool{}
	for _, id := range ids {
		listed[id] = true
	}
	if !listed[idle.ID] || listed[recent.ID] || listed[held.ID] || listed[fresh.ID] {
		t.Fatalf("listed = %v (idle %s, recent %s, held %s, fresh %s)", ids, idle.ID, recent.ID, held.ID, fresh.ID)
	}
}
