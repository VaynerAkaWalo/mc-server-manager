package server

import (
	"errors"
	"testing"
	"time"

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
	svc := CreateServerService(cl, nil)

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
	svc := CreateServerService(cl, nil)

	names, err := svc.ShutdownExpired()
	if err == nil || len(names) != 0 {
		t.Fatalf("names = %v, err = %v", names, err)
	}
}
