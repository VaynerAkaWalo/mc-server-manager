package server

import (
	"errors"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	serversv1alpha1 "github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type fakeCluster struct {
	items     []unstructured.Unstructured
	deleteErr error
	deleted   map[string]types.UID
}

func (f *fakeCluster) GetActiveServers() (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{Items: f.items}, nil
}

func (f *fakeCluster) DeployServerSpec(serversv1alpha1.McServerSpec) error { return nil }

func (f *fakeCluster) DeleteServer(name string, uid types.UID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted[name] = uid
	return nil
}

func serverResource(name string, age time.Duration) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"name":              name,
			"uid":               name + "-uid",
			"creationTimestamp": time.Now().Add(-age).UTC().Format(time.RFC3339),
		},
		"spec":   map[string]any{"expireAfter": int64(time.Hour / time.Millisecond)},
		"status": map[string]any{"startedTime": time.Now().Add(-age).UTC().Format(time.RFC3339)},
	}}
}

func TestShutdownExpiredDeletesByUID(t *testing.T) {
	cl := &fakeCluster{
		items:   []unstructured.Unstructured{serverResource("old-server", 3*time.Hour), serverResource("new-server", time.Minute)},
		deleted: map[string]types.UID{},
	}
	svc := CreateServerService(cl, &memoryStore{}, testVersions)

	names, err := svc.ShutdownExpired()
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 || names[0] != "old-server" || cl.deleted["old-server"] != "old-server-uid" {
		t.Fatalf("names = %v, deleted = %v", names, cl.deleted)
	}
}

func TestShutdownExpiredReportsFailedDeletes(t *testing.T) {
	cl := &fakeCluster{
		items:     []unstructured.Unstructured{serverResource("old-server", 3*time.Hour)},
		deleteErr: errors.New("forbidden"),
	}
	svc := CreateServerService(cl, &memoryStore{}, testVersions)

	names, err := svc.ShutdownExpired()
	if err == nil || len(names) != 0 {
		t.Fatalf("names = %v, err = %v", names, err)
	}
}

func idleRecord(name string, idleFor time.Duration) store.ServerRecord {
	record := testRecord(time.Now().Add(time.Hour))
	record.Name = name
	record.Status = "running"
	lastActive := time.Now().Add(-idleFor)
	record.LastActiveAt = &lastActive

	return record
}

func TestShutdownExpiredStopsIdleServers(t *testing.T) {
	cl := &fakeCluster{
		items:   []unstructured.Unstructured{serverResource("idle-server", time.Minute), serverResource("busy-server", time.Minute)},
		deleted: map[string]types.UID{},
	}
	db := &memoryStore{records: []store.ServerRecord{idleRecord("idle-server", 31*time.Minute), idleRecord("busy-server", time.Minute)}}
	svc := CreateServerService(cl, db, testVersions)

	names, err := svc.ShutdownExpired()
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 || names[0] != "idle-server" || cl.deleted["idle-server"] != "idle-server-uid" {
		t.Fatalf("names = %v, deleted = %v", names, cl.deleted)
	}
	if db.records[0].Status != "expired" || db.records[1].Status != "running" {
		t.Fatalf("statuses = %s, %s", db.records[0].Status, db.records[1].Status)
	}
}

func TestShutdownExpiredExpiresIdleServersMissingFromCluster(t *testing.T) {
	cl := &fakeCluster{deleted: map[string]types.UID{}}
	db := &memoryStore{records: []store.ServerRecord{idleRecord("gone-server", time.Hour)}}
	svc := CreateServerService(cl, db, testVersions)

	if _, err := svc.ShutdownExpired(); err != nil {
		t.Fatal(err)
	}

	if db.records[0].Status != "expired" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestShutdownExpiredKeepsIdleServersItFailedToDelete(t *testing.T) {
	cl := &fakeCluster{
		items:     []unstructured.Unstructured{serverResource("idle-server", time.Minute)},
		deleteErr: errors.New("forbidden"),
	}
	db := &memoryStore{records: []store.ServerRecord{idleRecord("idle-server", time.Hour)}}
	svc := CreateServerService(cl, db, testVersions)

	if _, err := svc.ShutdownExpired(); err == nil {
		t.Fatal("expected the failed delete to be reported")
	}

	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestShutdownExpiredStopsServersThatNeverStarted(t *testing.T) {
	cl := &fakeCluster{
		items:   []unstructured.Unstructured{serverResource("stuck-server", 61*time.Minute), serverResource("starting-server", time.Minute)},
		deleted: map[string]types.UID{},
	}
	stuck := idleRecord("stuck-server", 0)
	stuck.Status, stuck.LastActiveAt, stuck.CreatedAt = "new", nil, time.Now().Add(-61*time.Minute)
	starting := idleRecord("starting-server", 0)
	starting.Status, starting.LastActiveAt, starting.CreatedAt = "new", nil, time.Now().Add(-time.Minute)
	db := &memoryStore{records: []store.ServerRecord{stuck, starting}}
	svc := CreateServerService(cl, db, testVersions)

	names, err := svc.ShutdownExpired()
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 || names[0] != "stuck-server" || db.records[0].Status != "expired" || db.records[1].Status != "new" {
		t.Fatalf("names = %v, statuses = %s, %s", names, db.records[0].Status, db.records[1].Status)
	}
}
