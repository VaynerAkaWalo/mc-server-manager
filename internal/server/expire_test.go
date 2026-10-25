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

func serverResource(name string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"name": name, "uid": name + "-uid"},
	}}
}

func namedRecord(name string, status string) store.ServerRecord {
	record := testRecord(time.Now().Add(time.Hour))
	record.Name = name
	record.Status = status

	return record
}

func TestShutdownExpiredDeletesServersThatAreNoLongerActive(t *testing.T) {
	cl := &fakeCluster{
		items:   []unstructured.Unstructured{serverResource("expired-server"), serverResource("expiring-server"), serverResource("running-server")},
		deleted: map[string]types.UID{},
	}
	db := &memoryStore{records: []store.ServerRecord{
		namedRecord("expired-server", "expired"),
		namedRecord("expiring-server", "expiring"),
		namedRecord("running-server", "running"),
	}}
	svc := CreateServerService(cl, db, testVersions)

	names, err := svc.ShutdownExpired()
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 || names[0] != "expired-server" || cl.deleted["expired-server"] != "expired-server-uid" {
		t.Fatalf("names = %v, deleted = %v", names, cl.deleted)
	}
}

func TestShutdownExpiredDeletesServersPastTheirLifetime(t *testing.T) {
	cl := &fakeCluster{items: []unstructured.Unstructured{serverResource("old-server")}, deleted: map[string]types.UID{}}
	record := namedRecord("old-server", "running")
	record.ExpiresAt = time.Now().Add(-time.Minute)
	db := &memoryStore{records: []store.ServerRecord{record}}
	svc := CreateServerService(cl, db, testVersions)

	names, err := svc.ShutdownExpired()
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 || db.records[0].Status != "expired" {
		t.Fatalf("names = %v, status = %s", names, db.records[0].Status)
	}
}

func TestShutdownExpiredReportsFailedDeletes(t *testing.T) {
	cl := &fakeCluster{
		items:     []unstructured.Unstructured{serverResource("expired-server")},
		deleteErr: errors.New("forbidden"),
	}
	db := &memoryStore{records: []store.ServerRecord{namedRecord("expired-server", "expired")}}
	svc := CreateServerService(cl, db, testVersions)

	names, err := svc.ShutdownExpired()
	if err == nil || len(names) != 0 {
		t.Fatalf("names = %v, err = %v", names, err)
	}
}
