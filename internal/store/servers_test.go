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
