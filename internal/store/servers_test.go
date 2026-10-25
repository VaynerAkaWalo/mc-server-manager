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

	before, err := s.CountServers(ctx, ServerFilter{Status: "all"})
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

	after, err := s.CountServers(ctx, ServerFilter{Status: "all"})
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

	records, err := s.ListServers(ctx, ServerQuery{Filter: ServerFilter{Status: "active"}, Sort: "-createdAt", Limit: 20})
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

func TestServerCollectionPaginationAndFilters(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	ids := make(map[string]string)
	for _, status := range []string{"new", "running", "stopped", "expired"} {
		for i := 0; i < 3; i++ {
			record, err := s.CreateServer(ctx, NewServer{Name: "collection-" + status, Hostname: "test", Definition: map[string]any{}, ExpiresAt: now.Add(time.Hour)})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })
			ids[record.ID.String()] = status
			_, err = s.pool.Exec(ctx, "UPDATE mc_servers SET status = $1, created_at = $2 WHERE id = $3", status, now.Add(-time.Duration(i+1)*time.Hour), record.ID)
			if err != nil {
				t.Fatal(err)
			}
			if i == 2 {
				if _, err := s.pool.Exec(ctx, "UPDATE mc_servers SET expires_at = $1 WHERE id = $2", now.Add(-time.Hour), record.ID); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	for _, sort := range []string{"createdAt", "-createdAt", "expiresAt", "-expiresAt"} {
		t.Run(sort, func(t *testing.T) {
			query := ServerQuery{Filter: ServerFilter{Status: "all"}, Sort: sort, Limit: 2}
			seen := map[string]bool{}
			var previous *ServerRecord
			for page := 0; ; page++ {
				if page > 100 {
					t.Fatal("pagination did not terminate")
				}
				records, err := s.ListServers(ctx, query)
				if err != nil {
					t.Fatal(err)
				}
				hasMore := len(records) > query.Limit
				if hasMore {
					records = records[:query.Limit]
				}
				for _, record := range records {
					id := record.ID.String()
					if seen[id] {
						t.Fatalf("duplicate ID %s", id)
					}
					seen[id] = true
					if previous != nil {
						before, after := previous.CreatedAt, record.CreatedAt
						if sort == "expiresAt" || sort == "-expiresAt" {
							before, after = previous.ExpiresAt, record.ExpiresAt
						}
						descending := sort[0] == '-'
						if before.Equal(after) {
							if (previous.ID.String() > id) != descending {
								t.Fatal("incorrect ID tie-breaker")
							}
						} else if before.After(after) != descending {
							t.Fatal("incorrect timestamp order")
						}
					}
					copy := record
					previous = &copy
				}
				if !hasMore {
					break
				}
				last := records[len(records)-1]
				value := last.CreatedAt
				if sort == "expiresAt" || sort == "-expiresAt" {
					value = last.ExpiresAt
				}
				query.After = &ServerBoundary{Value: value, ID: last.ID}
			}
			for id := range ids {
				if !seen[id] {
					t.Fatalf("missing fixture %s", id)
				}
			}
		})
	}

	for _, status := range []string{"active", "all", "new", "running", "stopped", "expired"} {
		filter := ServerFilter{Status: status}
		records, err := s.ListServers(ctx, ServerQuery{Filter: filter, Sort: "-createdAt", Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		counts, err := s.CountServers(ctx, filter)
		if err != nil {
			t.Fatal(err)
		}
		if counts.Created != len(records) {
			t.Fatalf("%s stats %+v do not match %d records", status, counts, len(records))
		}
		for _, record := range records {
			if status != "active" && status != "all" && record.Status != status {
				t.Fatalf("%s filter returned %s", status, record.Status)
			}
			if status == "active" && record.Status != "running" && record.Status != "new" {
				t.Fatalf("active returned %s", record.Status)
			}
		}
	}
}
