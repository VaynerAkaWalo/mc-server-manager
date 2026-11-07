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

	record, err := s.CreateSave(context.Background(), owner, "world")
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

	played := time.Now().UTC().Truncate(time.Second)
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = $2 WHERE id = $1`, first.ID, played); err != nil {
		t.Fatal(err)
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
	restarted, err := serverOnSave(t, s, "lock-fourth", save)
	if err != nil {
		t.Fatalf("server after release: %v", err)
	}
	if restarted.LastPlayedAt == nil || !restarted.LastPlayedAt.Equal(played) {
		t.Fatalf("restarted world activity = %v", restarted.LastPlayedAt)
	}

	loaded, err := s.GetServer(ctx, restarted.ID)
	if err != nil || loaded.LastPlayedAt == nil || !loaded.LastPlayedAt.Equal(played) {
		t.Fatalf("loaded world activity = %v, %v", loaded.LastPlayedAt, err)
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

func TestSaveCreationHasNoCountCapAndListsOnlyTheOwner(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	owner := "many-saves-" + uuid.NewString()

	for i := 0; i < 4; i++ {
		newSave(t, s, owner)
	}
	newSave(t, s, owner+"-other")

	saves, err := s.ListSaves(ctx, owner)
	if err != nil || len(saves) != 4 {
		t.Fatalf("saves = %d, err = %v", len(saves), err)
	}
}

func TestConcurrentSaveCreationHasNoCountCap(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	owner := "race-" + uuid.NewString()
	t.Cleanup(func() { _, _ = s.pool.Exec(ctx, `DELETE FROM saves WHERE owner_subject = $1`, owner) })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.CreateSave(ctx, owner, "world")
		}()
	}
	wg.Wait()

	saves, err := s.ListSaves(ctx, owner)
	if err != nil || len(saves) != 8 {
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

func TestPersonalHistoryKeepsItsOwnerAfterTheWorldExpires(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	owner := uuid.NewString()
	save := newSave(t, s, owner)
	record, err := s.CreateServer(ctx, NewServer{Owner: owner, Name: "personal-" + owner, Hostname: "personal", Definition: map[string]any{"saveId": save.ID}, ExpiresAt: time.Now().Add(time.Hour), Save: &SaveUse{ID: save.ID, Owner: owner}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })
	if record.Owner != owner || !record.SaveAvailable || !record.SaveLocked {
		t.Fatalf("created = %+v", record)
	}
	played := time.Now().UTC().Truncate(time.Second)
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = $2 WHERE id = $1`, record.ID, played); err != nil {
		t.Fatal(err)
	}
	active, err := s.GetServer(ctx, record.ID)
	if err != nil || active.LastPlayedAt == nil || !active.LastPlayedAt.Equal(played) {
		t.Fatalf("world activity = %v, %v", active.LastPlayedAt, err)
	}
	if err := s.MarkStopped(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSave(ctx, save.ID, owner, func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}

	query := ServerQuery{Filter: ServerFilter{Status: "all", Owner: owner}, Sort: "-createdAt", Limit: 20}
	items, err := s.ListServers(ctx, query)
	if err != nil || len(items) != 1 || items[0].ID != record.ID || items[0].SaveAvailable || items[0].SaveLocked || items[0].LastPlayedAt == nil || !items[0].LastPlayedAt.Equal(played) {
		t.Fatalf("history = %+v, %v", items, err)
	}
	query.Filter.Owner = owner + "-other"
	if items, err := s.ListServers(ctx, query); err != nil || len(items) != 0 {
		t.Fatalf("foreign history = %+v, %v", items, err)
	}
	counts, err := s.CountServers(ctx, ServerFilter{Status: "all", Owner: owner})
	if err != nil || counts.Created != 1 {
		t.Fatalf("counts = %+v, %v", counts, err)
	}
}

func TestStoppedWorldReservesItsNameUntilTheSaveIsDeleted(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, uuid.NewString())
	name := "reserved-" + uuid.NewString()
	record, err := serverOnSave(t, s, name, save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, record.ID); err != nil {
		t.Fatal(err)
	}

	guest := NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}
	if taken, err := s.CreateServer(ctx, guest); !errors.Is(err, ErrActiveNameTaken) {
		if err == nil {
			_ = s.DeleteServer(ctx, taken.ID)
		}
		t.Fatalf("Guest taking a saved name = %v", err)
	}
	other := newSave(t, s, save.Owner)
	if _, err := serverOnSave(t, s, name, other); !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("another world taking a saved name = %v", err)
	}

	resumed, err := serverOnSave(t, s, name, save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, resumed.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, resumed.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSave(ctx, save.ID, save.Owner, func(context.Context) error { return errors.New("bucket unavailable") }); err == nil {
		t.Fatal("expected purge failure")
	}
	if taken, err := s.CreateServer(ctx, guest); !errors.Is(err, ErrActiveNameTaken) {
		if err == nil {
			_ = s.DeleteServer(ctx, taken.ID)
		}
		t.Fatalf("name after failed purge = %v", err)
	}
	if err := s.DeleteSave(ctx, save.ID, save.Owner, func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	available, err := s.CreateServer(ctx, guest)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, available.ID) })
}

func TestIdleWorldExpiryReleasesItsReservedName(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, uuid.NewString())
	name := "idle-reserved-" + uuid.NewString()
	record, err := serverOnSave(t, s, name, save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, record.ID); err != nil {
		t.Fatal(err)
	}
	if taken, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, ErrActiveNameTaken) {
		if err == nil {
			_ = s.DeleteServer(ctx, taken.ID)
		}
		t.Fatalf("name before idle expiry = %v", err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE saves SET created_at = now() - interval '91 days' WHERE id = $1`, save.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteIdleSave(ctx, save.ID, 90*24*time.Hour, func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	available, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, available.ID) })
}
