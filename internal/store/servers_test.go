package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"slices"
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

	create := func(name string, expiresIn time.Duration, status string) ServerRecord {
		t.Helper()
		record, err := s.CreateServer(ctx, NewServer{Name: name, Hostname: name, Definition: map[string]any{}, ExpiresAt: time.Now().Add(expiresIn)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })
		if _, err := s.pool.Exec(ctx, "UPDATE mc_servers SET status = $1 WHERE id = $2", status, record.ID); err != nil {
			t.Fatal(err)
		}
		return record
	}

	active := create("list-active", time.Hour, "running")
	oldRun := create("list-reused", -2*time.Hour, "expired")
	latestRun := create("list-reused", -time.Hour, "expired")
	gone := create("list-gone", -time.Hour, "new")

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
			record, err := s.CreateServer(ctx, NewServer{Name: fmt.Sprintf("collection-%s-%d", status, i), Hostname: "test", Definition: map[string]any{}, ExpiresAt: now.Add(time.Hour)})
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
					if stored, ok := ids[id]; ok && record.Status != stored {
						t.Fatalf("status = %s, want stored %s", record.Status, stored)
					}
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
			if status == "active" && !record.ExpiresAt.After(time.Now()) {
				t.Fatalf("active returned expired record %s", record.ID)
			}
			if status == "active" && record.Status != "running" && record.Status != "new" {
				t.Fatalf("active returned %s", record.Status)
			}
		}
	}
}

func TestActiveNamesAreUnique(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	input := NewServer{Name: "unique-name", Hostname: "u", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}

	first, err := s.CreateServer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, first.ID) })

	if _, err := s.CreateServer(ctx, input); !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("err = %v", err)
	}

	if err := s.MarkStopped(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateServer(ctx, input)
	if err != nil {
		t.Fatalf("name was not released: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, second.ID) })

	stopped, err := s.GetServer(ctx, first.ID)
	if err != nil || stopped.Status != "stopped" {
		t.Fatalf("status = %q, err = %v", stopped.Status, err)
	}
	if _, err := s.GetServer(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestMarkExpiredReleasesName(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	old, err := s.CreateServer(ctx, NewServer{Name: "expiring-name", Hostname: "e", Definition: map[string]any{}, ExpiresAt: time.Now().Add(-time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, old.ID) })

	if err := s.MarkExpired(ctx); err != nil {
		t.Fatal(err)
	}

	record, err := s.GetServer(ctx, old.ID)
	if err != nil || record.Status != "expired" {
		t.Fatalf("status = %q, err = %v", record.Status, err)
	}
}

func TestMonitoringTransitions(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()

	record, err := s.CreateServer(ctx, NewServer{Name: "monitored", Hostname: "a", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteServer(ctx, record.ID) })

	statusOf := func() string {
		stored, err := s.GetServer(ctx, record.ID)
		if err != nil {
			t.Fatal(err)
		}
		return stored.Status
	}

	if err := s.MarkDegraded(ctx, record.ID); err != nil || statusOf() != "new" {
		t.Fatalf("a new server cannot be degraded, status = %s, err = %v", statusOf(), err)
	}
	if err := s.MarkRunning(ctx, record.ID); err != nil || statusOf() != "running" {
		t.Fatalf("status = %s, err = %v", statusOf(), err)
	}
	if err := s.MarkDegraded(ctx, record.ID); err != nil || statusOf() != "degraded" {
		t.Fatalf("status = %s, err = %v", statusOf(), err)
	}

	monitored, err := s.ListMonitoredServers(ctx)
	if err != nil || !slices.ContainsFunc(monitored, func(r ServerRecord) bool { return r.ID == record.ID }) {
		t.Fatalf("a degraded server is still monitored, got %v, err = %v", monitored, err)
	}

	if _, err := s.CreateServer(ctx, NewServer{Name: "monitored", Hostname: "a", Definition: map[string]any{}, ExpiresAt: time.Now().Add(time.Hour)}); !errors.Is(err, ErrActiveNameTaken) {
		t.Fatalf("a degraded server keeps its name, err = %v", err)
	}

	if err := s.MarkStopped(ctx, record.ID); err != nil || statusOf() != "stopped" {
		t.Fatalf("status = %s, err = %v", statusOf(), err)
	}
	if err := s.MarkRunning(ctx, record.ID); err != nil || statusOf() != "stopped" {
		t.Fatalf("a stopped server cannot be revived, status = %s, err = %v", statusOf(), err)
	}
}
