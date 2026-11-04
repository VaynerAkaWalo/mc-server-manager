package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/auth"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (s *memoryStore) GetSave(_ context.Context, id uuid.UUID, owner string) (store.SaveRecord, error) {
	if s.saves[id] != owner || owner == "" {
		return store.SaveRecord{}, store.ErrSaveNotFound
	}
	return store.SaveRecord{ID: id, Owner: owner}, nil
}

func (s *memoryStore) LatestSaveDefinition(_ context.Context, id uuid.UUID) (json.RawMessage, bool, error) {
	definition, found := s.played[id]
	return definition, found, nil
}

func (s *memoryStore) ListReleasableSaveLocks(context.Context) ([]store.ServerRecord, error) {
	released := make([]store.ServerRecord, 0)
	for _, record := range s.records {
		if _, held := s.lockNames[record.ID]; held && (record.Status == "stopped" || record.Status == "expired") {
			released = append(released, record)
		}
	}
	return released, nil
}

func (s *memoryStore) ReleaseSaveLock(_ context.Context, id uuid.UUID) error {
	for save, server := range s.locks {
		if server == id {
			delete(s.locks, save)
		}
	}
	delete(s.lockNames, id)
	return nil
}

var twoVersions = definition.VersionSet{
	definition.ServerType: {LatestRelease: "1.21.10", Items: []definition.Version{
		{ID: "1.21.10", Kind: definition.KindRelease},
		{ID: "1.21.3", Kind: definition.KindRelease},
	}},
}

func savedGame(owner string) (*memoryStore, uuid.UUID) {
	id := uuid.New()
	return &memoryStore{saves: map[uuid.UUID]string{id: owner}}, id
}

func freeRequest(saveID uuid.UUID, version string) server.Request {
	return server.Request{Name: "my-server", Tier: "free", SaveID: saveID.String(), Options: map[string]string{"VERSION": version}}
}

func TestProvisionWithSaveDeploysSaveSpec(t *testing.T) {
	db, saveID := savedGame("alice")
	cl := &fakeCluster{}
	svc := CreateServerService(cl, db, twoVersions)

	created, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.10"))
	if err != nil {
		t.Fatal(err)
	}

	if len(cl.deployed) != 1 || cl.deployed[0].Save == nil || cl.deployed[0].Save.ID != saveID.String() {
		t.Fatalf("deployed = %+v", cl.deployed)
	}
	if created.SaveID != saveID.String() {
		t.Fatalf("created = %+v", created)
	}
}

func TestProvisionWithoutSaveDeploysNoSaveSpec(t *testing.T) {
	cl := &fakeCluster{}
	svc := CreateServerService(cl, &memoryStore{}, twoVersions)

	_, err := svc.provisionServer(context.Background(), "", server.Request{Name: "my-server", Tier: "guest", Options: map[string]string{"VERSION": "1.21.10"}})
	if err != nil {
		t.Fatal(err)
	}

	if cl.deployed[0].Save != nil {
		t.Fatalf("deployed = %+v", cl.deployed)
	}
}

func TestProvisionRejectsSaveRequests(t *testing.T) {
	db, saveID := savedGame("alice")
	db.locks = map[uuid.UUID]uuid.UUID{}
	svc := CreateServerService(&fakeCluster{}, db, twoVersions)

	cases := map[string]struct {
		owner   string
		request server.Request
		want    error
	}{
		"free tier without login": {"", freeRequest(saveID, "1.21.10"), ErrLoginRequired},
		"other owner":             {"mallory", freeRequest(saveID, "1.21.10"), ErrSaveNotFound},
		"unknown save":            {"alice", freeRequest(uuid.New(), "1.21.10"), ErrSaveNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.provisionServer(context.Background(), tc.owner, tc.request); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}

	guest := freeRequest(saveID, "1.21.10")
	guest.Tier = "guest"
	var invalid definition.InvalidDefinitionError
	if _, err := svc.provisionServer(context.Background(), "alice", guest); !errors.As(err, &invalid) {
		t.Fatalf("guest tier with a save: %v", err)
	}

	malformed := freeRequest(saveID, "1.21.10")
	malformed.SaveID = "not-a-uuid"
	if _, err := svc.provisionServer(context.Background(), "alice", malformed); !errors.As(err, &invalid) {
		t.Fatalf("malformed save id: %v", err)
	}
}

func TestProvisionRejectsSaveThatAnotherServerHolds(t *testing.T) {
	db, saveID := savedGame("alice")
	svc := CreateServerService(&fakeCluster{}, db, twoVersions)

	if _, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.10")); err != nil {
		t.Fatal(err)
	}
	second := freeRequest(saveID, "1.21.10")
	second.Name = "other-server"
	if _, err := svc.provisionServer(context.Background(), "alice", second); !errors.Is(err, ErrSaveInUse) {
		t.Fatalf("err = %v", err)
	}
}

func TestProvisionRefusesToDowngradeASave(t *testing.T) {
	db, saveID := savedGame("alice")
	db.played = map[uuid.UUID]json.RawMessage{saveID: json.RawMessage(`{"options":{"VERSION":"1.21.10"}}`)}
	svc := CreateServerService(&fakeCluster{}, db, twoVersions)

	var invalid definition.InvalidDefinitionError
	if _, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.3")); !errors.As(err, &invalid) || !strings.Contains(err.Error(), "1.21.10") {
		t.Fatalf("err = %v", err)
	}
	if _, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.10")); err != nil {
		t.Fatalf("same version: %v", err)
	}
}

func TestProvisionAllowsUpgradingASave(t *testing.T) {
	db, saveID := savedGame("alice")
	db.played = map[uuid.UUID]json.RawMessage{saveID: json.RawMessage(`{"options":{"VERSION":"1.21.3"}}`)}
	svc := CreateServerService(&fakeCluster{}, db, twoVersions)

	if _, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.10")); err != nil {
		t.Fatal(err)
	}
}

func TestSaveLockIsKeptWhileTheClusterResourceStillExists(t *testing.T) {
	db, saveID := savedGame("alice")
	cl := &fakeCluster{}
	svc := CreateServerService(cl, db, twoVersions)
	created, err := svc.provisionServer(context.Background(), "alice", freeRequest(saveID, "1.21.10"))
	if err != nil {
		t.Fatal(err)
	}
	db.setStatus(created.ID, []string{"new"}, "expired")

	cl.items = []unstructured.Unstructured{serverResource("my-server")}
	svc.releaseSaveLocks(context.Background())
	if _, held := db.locks[saveID]; !held {
		t.Fatal("the lock was released while the server was still shutting down")
	}

	cl.items = nil
	svc.releaseSaveLocks(context.Background())
	if _, held := db.locks[saveID]; held {
		t.Fatal("the lock was not released after the server was gone")
	}
}

type fakeIdentity struct {
	session store.Session
	err     error
	origin  bool
}

func (f fakeIdentity) Identify(http.ResponseWriter, *http.Request) (store.Session, error) {
	return f.session, f.err
}

func (f fakeIdentity) SameOrigin(*http.Request) bool { return f.origin }

func provision(t *testing.T, identity fakeIdentity, body string) *httptest.ResponseRecorder {
	t.Helper()

	svc := CreateServerService(&fakeCluster{}, &memoryStore{}, twoVersions)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", strings.NewReader(body))
	if err := NewServerHandlers(svc, identity).provisionServer(rec, req); err != nil {
		t.Fatal(err)
	}
	return rec
}

func TestProvisionMapsAccountFailures(t *testing.T) {
	free := `{"name":"my-server","tier":"free","options":{"VERSION":"1.21.10"}}`

	if rec := provision(t, fakeIdentity{err: auth.ErrNoSession}, free); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous free tier: %d %s", rec.Code, rec.Body)
	}
	if rec := provision(t, fakeIdentity{session: store.Session{Subject: "alice"}}, free); rec.Code != http.StatusForbidden {
		t.Fatalf("foreign origin: %d %s", rec.Code, rec.Body)
	}
	if rec := provision(t, fakeIdentity{session: store.Session{Subject: "alice"}, origin: true}, free); rec.Code != http.StatusCreated {
		t.Fatalf("free tier: %d %s", rec.Code, rec.Body)
	}

	guest := `{"name":"my-server","tier":"guest","options":{"VERSION":"1.21.10"}}`
	if rec := provision(t, fakeIdentity{err: auth.ErrNoSession}, guest); rec.Code != http.StatusCreated {
		t.Fatalf("anonymous guest: %d %s", rec.Code, rec.Body)
	}
	if rec := provision(t, fakeIdentity{session: store.Session{Subject: "alice"}}, guest); rec.Code != http.StatusCreated {
		t.Fatalf("guest tier does not use the session, so the origin is not checked: %d %s", rec.Code, rec.Body)
	}
}
