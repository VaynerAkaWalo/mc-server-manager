package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func connectTestStore(t *testing.T) *Store {
	t.Helper()

	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Skip("TEST_DB_URL is not set")
	}

	s, err := Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)

	if err := s.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	return s
}

func TestMigrateIsRepeatable(t *testing.T) {
	s := connectTestStore(t)

	if err := s.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndDeleteServer(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	record, err := s.CreateServer(ctx, NewServer{
		Name:       "test-server",
		Hostname:   "test-server.mc-server.blamedevs.com",
		Definition: map[string]any{"tier": "small"},
		ExpiresAt:  time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.ID.String() == "" || record.CreatedAt.IsZero() {
		t.Fatalf("expected generated id and creation time, got %+v", record)
	}

	var tier string
	err = s.pool.QueryRow(ctx, `SELECT definition->>'tier' FROM mc_servers WHERE id = $1`, record.ID).Scan(&tier)
	if err != nil || tier != "small" {
		t.Fatalf("expected stored definition with tier small, got %q, %v", tier, err)
	}

	if err := s.DeleteServer(ctx, record.ID); err != nil {
		t.Fatal(err)
	}

	var count int
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM mc_servers WHERE id = $1`, record.ID).Scan(&count)
	if count != 0 {
		t.Fatalf("expected record to be deleted, found %d", count)
	}
}

func TestCountServers(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	before, err := s.CountServers(ctx)
	if err != nil {
		t.Fatal(err)
	}

	active, err := s.CreateServer(ctx, NewServer{Name: "count-active", Hostname: "a", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	expired, err := s.CreateServer(ctx, NewServer{Name: "count-expired", Hostname: "b", Definition: map[string]any{}, ExpiresAt: time.Now().Add(-time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = s.DeleteServer(ctx, active.ID)
		_ = s.DeleteServer(ctx, expired.ID)
	})

	after, err := s.CountServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.Online != before.Online+1 || after.Created != before.Created+2 {
		t.Fatalf("expected one more online and two more created, got before %+v after %+v", before, after)
	}
}

func TestListServers(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	create := func(name string, expiresIn time.Duration) ServerRecord {
		t.Helper()
		record, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(expiresIn)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })
		return record
	}

	active := create("list-active", time.Hour)
	oldRun := create("list-reused", -2*time.Hour)
	latestRun := create("list-reused", -time.Hour)
	gone := create("list-gone", -time.Hour)

	records, err := s.ListServers(ctx)
	if err != nil {
		t.Fatal(err)
	}

	listed := map[string]bool{}
	for _, record := range records {
		listed[record.ID.String()] = true
	}
	if !listed[active.ID.String()] || listed[latestRun.ID.String()] || listed[oldRun.ID.String()] || listed[gone.ID.String()] {
		t.Fatalf("unexpected records: %v", listed)
	}

	for i := 1; i < len(records); i++ {
		if records[i-1].CreatedAt.Before(records[i].CreatedAt) {
			t.Fatalf("records are not newest first")
		}
	}
}
