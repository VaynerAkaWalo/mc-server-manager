package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
)

type memoryStore struct {
	records []store.ServerRecord
}

func (s *memoryStore) CreateServer(_ context.Context, input store.NewServer) (store.ServerRecord, error) {
	definition, err := json.Marshal(input.Definition)
	if err != nil {
		return store.ServerRecord{}, err
	}

	record := store.ServerRecord{ID: uuid.New(), Name: input.Name, Hostname: input.Hostname, Definition: definition, CreatedAt: time.Now(), ExpiresAt: input.ExpiresAt, Status: "new"}
	s.records = append([]store.ServerRecord{record}, s.records...)
	return record, nil
}

func (s *memoryStore) DeleteServer(context.Context, uuid.UUID) error { return nil }
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
	if len(listed.Items) != 1 || listed.Items[0].ID != record.ID || listed.Items[0].Status != server.StatusNew || !listed.Items[0].ExpiresAt.Equal(record.ExpiresAt) {
		t.Fatalf("listed = %+v", listed)
	}
}
