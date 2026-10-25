package server

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type memoryStore struct {
	records    []store.ServerRecord
	expireRuns int
}

func (s *memoryStore) CreateServer(_ context.Context, input store.NewServer) (store.ServerRecord, error) {
	definition, err := json.Marshal(input.Definition)
	if err != nil {
		return store.ServerRecord{}, err
	}

	for _, existing := range s.records {
		if existing.Name == input.Name && isActive(existing.Status) {
			return store.ServerRecord{}, store.ErrActiveNameTaken
		}
	}

	record := store.ServerRecord{ID: uuid.New(), Name: input.Name, Hostname: input.Hostname, Definition: definition, CreatedAt: time.Now(), ExpiresAt: input.ExpiresAt, Status: "new"}
	s.records = append([]store.ServerRecord{record}, s.records...)
	return record, nil
}

func (s *memoryStore) DeleteServer(context.Context, uuid.UUID) error { return nil }

func (s *memoryStore) GetServer(_ context.Context, id uuid.UUID) (store.ServerRecord, error) {
	for _, record := range s.records {
		if record.ID == id {
			return record, nil
		}
	}
	return store.ServerRecord{}, store.ErrNotFound
}

func (s *memoryStore) setStatus(id uuid.UUID, from []string, to string) {
	for i := range s.records {
		if s.records[i].ID == id && slices.Contains(from, s.records[i].Status) {
			s.records[i].Status = to
		}
	}
}

func (s *memoryStore) ListMonitoredServers(context.Context) ([]store.ServerRecord, error) {
	monitored := make([]store.ServerRecord, 0)
	for _, record := range s.records {
		if isActive(record.Status) {
			monitored = append(monitored, record)
		}
	}
	return monitored, nil
}

func (s *memoryStore) MarkRunning(_ context.Context, id uuid.UUID) error {
	s.setStatus(id, []string{"new", "degraded"}, "running")
	return nil
}

func (s *memoryStore) MarkDegraded(_ context.Context, id uuid.UUID) error {
	s.setStatus(id, []string{"running"}, "degraded")
	return nil
}

func (s *memoryStore) MarkStopped(_ context.Context, id uuid.UUID) error {
	for i := range s.records {
		if s.records[i].ID == id {
			s.records[i].Status = "stopped"
		}
	}
	return nil
}

func (s *memoryStore) MarkExpired(context.Context) error {
	s.expireRuns++
	return nil
}
func (s *memoryStore) CountServers(context.Context, store.ServerFilter) (store.ServerCounts, error) {
	return store.ServerCounts{}, nil
}
func (s *memoryStore) ListServers(context.Context, store.ServerQuery) ([]store.ServerRecord, error) {
	return s.records, nil
}

func TestProvisionAndListUseSameRepresentation(t *testing.T) {
	cl := &fakeCluster{}
	db := &memoryStore{}
	svc := CreateServerService(cl, db)
	request := server.Request{Name: "my-server", DurationMs: time.Hour.Milliseconds(), Tier: "wooden", Options: map[string]string{"MOTD": "hello", "VERSION": "1.21.3"}}

	created, err := svc.provisionServer(request)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := svc.listServers(context.Background(), store.ServerQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("listed = %+v", listed)
	}

	createdJSON, _ := json.Marshal(created)
	listedJSON, _ := json.Marshal(listed.Items[0])
	if string(createdJSON) != string(listedJSON) {
		t.Fatalf("created %s differs from listed %s", createdJSON, listedJSON)
	}
	if created.ID == uuid.Nil || created.Options["MOTD"] != "hello" || len(created.Options) != 2 {
		t.Fatalf("created = %+v", created)
	}
	if created.ExpiresAt.Sub(created.CreatedAt).Round(time.Second) != time.Hour {
		t.Fatalf("durationMs did not set expiry: %+v", created)
	}
}

func TestListUsesOnlyDatabase(t *testing.T) {
	record := testRecord(time.Now().Add(time.Hour))
	record.Status = "new"
	db := &memoryStore{records: []store.ServerRecord{record}}
	svc := CreateServerService(nil, db)

	listed, err := svc.listServers(context.Background(), store.ServerQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 1 || listed.Items[0].ID != record.ID || listed.Items[0].Status != server.StatusNew {
		t.Fatalf("listed = %+v", listed)
	}
	if !listed.Items[0].ExpiresAt.Equal(record.ExpiresAt) {
		t.Fatalf("listed = %+v", listed)
	}
}

func TestProvisionRejectsActiveName(t *testing.T) {
	svc := CreateServerService(&fakeCluster{}, &memoryStore{})
	request := server.Request{Name: "my-server", DurationMs: time.Hour.Milliseconds(), Tier: "wooden"}

	if _, err := svc.provisionServer(request); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.provisionServer(request); !errors.Is(err, ErrNameInUse) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetServerMissing(t *testing.T) {
	svc := CreateServerService(nil, &memoryStore{})

	if _, err := svc.getServer(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestDeleteServerStopsClusterResourceAndMarksRecord(t *testing.T) {
	record := testRecord(time.Now().Add(time.Hour))
	record.Status = "running"
	db := &memoryStore{records: []store.ServerRecord{record}}
	cl := &fakeCluster{items: []unstructured.Unstructured{serverResource("my-server", time.Minute)}, deleted: map[string]types.UID{}}
	svc := CreateServerService(cl, db)

	if err := svc.deleteServer(context.Background(), record.ID); err != nil {
		t.Fatal(err)
	}
	if cl.deleted["my-server"] != "my-server-uid" || db.records[0].Status != "stopped" {
		t.Fatalf("deleted = %v, status = %s", cl.deleted, db.records[0].Status)
	}

	if err := svc.deleteServer(context.Background(), record.ID); err != nil {
		t.Fatalf("repeated delete: %v", err)
	}
	if err := svc.deleteServer(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
}

func TestDeleteServerKeepsRecordWhenClusterFails(t *testing.T) {
	record := testRecord(time.Now().Add(time.Hour))
	record.Status = "running"
	db := &memoryStore{records: []store.ServerRecord{record}}
	cl := &fakeCluster{items: []unstructured.Unstructured{serverResource("my-server", time.Minute)}, deleteErr: errors.New("forbidden")}

	svc := CreateServerService(cl, db)
	if err := svc.deleteServer(context.Background(), record.ID); err == nil {
		t.Fatal("expected error")
	}
	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}
