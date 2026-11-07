package store

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func nameMigrationStore(t *testing.T) *Store {
	t.Helper()
	adminURL := os.Getenv("TEST_DB_URL")
	if adminURL == "" {
		t.Skip("TEST_DB_URL is not set")
	}
	ctx := context.Background()
	admin, err := Connect(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	database := "migrate_names_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.pool.Exec(ctx, "CREATE DATABASE "+database); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = admin.pool.Exec(ctx, "DROP DATABASE "+database+" WITH (FORCE)") })
	parsed, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + database
	s, err := Connect(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	db := stdlib.OpenDBFromPool(s.pool)
	t.Cleanup(func() { _ = db.Close() })
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationsFS())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 9); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReservationMigrationProtectsExistingWorldNames(t *testing.T) {
	s := nameMigrationStore(t)
	ctx := context.Background()
	save := newSave(t, s, "alice")
	first, err := serverOnSave(t, s, "existing-world", save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	second, err := serverOnSave(t, s, "existing-world", save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateServer(ctx, NewServer{Name: "existing-world", Hostname: "existing-world", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("taking existing saved name = %v", err)
	}
}

func TestReservationMigrationRejectsAmbiguousNamesWithoutChangingWorlds(t *testing.T) {
	s := nameMigrationStore(t)
	ctx := context.Background()
	for _, owner := range []string{"alice", "bob"} {
		save := newSave(t, s, owner)
		record, err := serverOnSave(t, s, "shared-name", save)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.MarkStopped(ctx, record.ID); err != nil {
			t.Fatal(err)
		}
		if err := s.ReleaseSaveLock(ctx, record.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Migrate(ctx); err == nil {
		t.Fatal("migration must reject a name shared by two live saves")
	}
	var reservationTable *string
	if err := s.pool.QueryRow(ctx, "SELECT to_regclass('saved_server_names')::text").Scan(&reservationTable); err != nil {
		t.Fatal(err)
	}
	if reservationTable != nil {
		t.Fatal("failed migration left the reservation table")
	}
	var version int
	if err := s.pool.QueryRow(ctx, "SELECT max(version_id) FROM goose_db_version WHERE is_applied").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 9 {
		t.Fatalf("version after failed migration = %d", version)
	}
	var worlds int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM saves").Scan(&worlds); err != nil {
		t.Fatal(err)
	}
	if worlds != 2 {
		t.Fatalf("worlds after failed migration = %d", worlds)
	}
}

func TestGuestInsertWaitsForAnUncommittedSavedName(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	save := newSave(t, s, uuid.NewString())
	name := "concurrent-reserved-" + uuid.NewString()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var holderPID int
	if err := tx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&holderPID); err != nil {
		t.Fatal(err)
	}

	var id uuid.UUID
	if err := tx.QueryRow(ctx, `INSERT INTO mc_servers (name, hostname, definition, expires_at, save_id, status) VALUES ($1, $1, '{}', now(), $2, 'stopped') RETURNING id`, name, save.ID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, id) })
	result := make(chan error, 1)
	go func() {
		record, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
		if err == nil {
			_ = s.DeleteServer(ctx, record.ID)
		}
		result <- err
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		select {
		case err := <-result:
			t.Fatalf("Guest insert finished before the saved name committed: %v", err)
		default:
		}
		var waiting bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE wait_event = 'advisory' AND $1 = ANY(pg_blocking_pids(pid)))`, holderPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Guest insert did not wait for the name lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-result; !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("taking concurrent saved name = %v", err)
	}
}

func TestReservationMigrationDoesNotStopAnExistingGuestUsingASavedName(t *testing.T) {
	s := nameMigrationStore(t)
	ctx := context.Background()
	save := newSave(t, s, "alice")
	saved, err := serverOnSave(t, s, "guest-overlap", save)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkStopped(ctx, saved.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSaveLock(ctx, saved.ID); err != nil {
		t.Fatal(err)
	}
	guest, err := s.CreateServer(ctx, NewServer{Name: saved.Name, Hostname: saved.Hostname, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, guest.ID) })
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	running, err := s.GetServer(ctx, guest.ID)
	if err != nil || running.Status != "new" {
		t.Fatalf("existing Guest = %+v, %v", running, err)
	}
	if err := s.MarkStopped(ctx, guest.ID); err != nil {
		t.Fatal(err)
	}
	if taken, err := s.CreateServer(ctx, NewServer{Name: saved.Name, Hostname: saved.Hostname, Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, ErrActiveNameTaken) {
		if err == nil {
			_ = s.DeleteServer(ctx, taken.ID)
		}
		t.Fatalf("Guest reusing migrated saved name = %v", err)
	}
	if _, err := serverOnSave(t, s, saved.Name, save); err != nil {
		t.Fatal(err)
	}
}
