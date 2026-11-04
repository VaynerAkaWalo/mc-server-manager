package save

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/auth"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/google/uuid"
)

type memorySaves struct {
	records  []store.SaveRecord
	idle     []uuid.UUID
	inUse    map[uuid.UUID]bool
	revived  map[uuid.UUID]bool
	purgeErr error
	purged   []string
	idleFor  time.Duration
}

func (m *memorySaves) CreateSave(_ context.Context, owner, name string, limit int) (store.SaveRecord, error) {
	owned := 0
	for _, record := range m.records {
		if record.Owner == owner {
			owned++
		}
	}
	if owned >= limit {
		return store.SaveRecord{}, store.ErrSaveLimit
	}
	record := store.SaveRecord{ID: uuid.New(), Owner: owner, Name: name, CreatedAt: time.Now()}
	m.records = append(m.records, record)
	return record, nil
}

func (m *memorySaves) ListSaves(_ context.Context, owner string) ([]store.SaveRecord, error) {
	owned := make([]store.SaveRecord, 0)
	for _, record := range m.records {
		if record.Owner == owner {
			owned = append(owned, record)
		}
	}
	return owned, nil
}

func (m *memorySaves) RenameSave(_ context.Context, id uuid.UUID, owner, name string) (store.SaveRecord, error) {
	for i := range m.records {
		if m.records[i].ID == id && m.records[i].Owner == owner {
			m.records[i].Name = name
			return m.records[i], nil
		}
	}
	return store.SaveRecord{}, store.ErrSaveNotFound
}

func (m *memorySaves) DeleteIdleSave(ctx context.Context, id uuid.UUID, idle time.Duration, purge func(context.Context) error) error {
	if m.revived[id] {
		return store.ErrSaveInUse
	}
	return m.DeleteSave(ctx, id, "", purge)
}

func (m *memorySaves) DeleteSave(ctx context.Context, id uuid.UUID, owner string, purge func(context.Context) error) error {
	if m.inUse[id] {
		return store.ErrSaveInUse
	}
	for i, record := range m.records {
		if record.ID != id || (owner != "" && record.Owner != owner) {
			continue
		}
		if err := purge(ctx); err != nil {
			return err
		}
		m.records = append(m.records[:i], m.records[i+1:]...)
		return nil
	}
	return store.ErrSaveNotFound
}

func (m *memorySaves) ListSavesIdleSince(_ context.Context, idle time.Duration) ([]uuid.UUID, error) {
	m.idleFor = idle
	return m.idle, nil
}

func (m *memorySaves) Purge(_ context.Context, prefix string) error {
	if m.purgeErr != nil {
		return m.purgeErr
	}
	m.purged = append(m.purged, prefix)
	return nil
}

type fixedSession struct{ subject string }

func (f fixedSession) Authenticated(next auth.SessionHandler) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return next(w, r, store.Session{Subject: f.subject})
	}
}

func call(t *testing.T, saves *memorySaves, subject, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	router := newTestRouter(NewHandlers(NewService(saves, saves), fixedSession{subject}))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestCreateListRenameAndDeleteASave(t *testing.T) {
	saves := &memorySaves{}

	created := call(t, saves, "alice", "POST", "/api/v1/saves", `{"name":"  Survival world "}`)
	var save Response
	if err := json.Unmarshal(created.Body.Bytes(), &save); err != nil || created.Code != http.StatusCreated || save.Name != "Survival world" {
		t.Fatalf("create: %d %s", created.Code, created.Body)
	}

	listed := call(t, saves, "alice", "GET", "/api/v1/saves", "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), save.ID.String()) || !strings.Contains(listed.Body.String(), `"nextCursor":""`) {
		t.Fatalf("list: %d %s", listed.Code, listed.Body)
	}
	if other := call(t, saves, "bob", "GET", "/api/v1/saves", ""); strings.Contains(other.Body.String(), save.ID.String()) {
		t.Fatalf("another user sees the save: %s", other.Body)
	}
	if page := call(t, saves, "alice", "GET", "/api/v1/saves?limit=1&sort=-createdAt", ""); page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `"nextCursor":""`) || strings.Count(page.Body.String(), `"id"`) != 1 {
		t.Fatalf("page: %d %s", page.Code, page.Body)
	}

	renamed := call(t, saves, "alice", "PATCH", "/api/v1/saves/"+save.ID.String(), `{"name":"Creative"}`)
	if renamed.Code != http.StatusOK || !strings.Contains(renamed.Body.String(), `"name":"Creative"`) {
		t.Fatalf("rename: %d %s", renamed.Code, renamed.Body)
	}

	if rec := call(t, saves, "bob", "DELETE", "/api/v1/saves/"+save.ID.String(), ""); rec.Code != http.StatusNotFound {
		t.Fatalf("delete by another user: %d", rec.Code)
	}
	if rec := call(t, saves, "alice", "DELETE", "/api/v1/saves/"+save.ID.String(), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body)
	}
	if len(saves.purged) != 1 || saves.purged[0] != "saves/"+save.ID.String() {
		t.Fatalf("purged = %v", saves.purged)
	}
}

func TestSaveRequestsAreRejected(t *testing.T) {
	saves := &memorySaves{}
	for i := 0; i < MaxSavesPerUser; i++ {
		call(t, saves, "alice", "POST", "/api/v1/saves", `{"name":"world"}`)
	}
	id := saves.records[0].ID.String()

	cases := map[string]struct {
		method, path, body string
		status             int
	}{
		"limit reached":      {"POST", "/api/v1/saves", `{"name":"fourth"}`, http.StatusUnprocessableEntity},
		"empty name":         {"PATCH", "/api/v1/saves/" + id, `{"name":"   "}`, http.StatusUnprocessableEntity},
		"name too long":      {"PATCH", "/api/v1/saves/" + id, `{"name":"` + strings.Repeat("a", MaxNameLength+1) + `"}`, http.StatusUnprocessableEntity},
		"control characters": {"PATCH", "/api/v1/saves/" + id, `{"name":"a\u0000b"}`, http.StatusUnprocessableEntity},
		"malformed body":     {"POST", "/api/v1/saves", `{nope`, http.StatusBadRequest},
		"trailing data":      {"POST", "/api/v1/saves", `{"name":"x"} {}`, http.StatusBadRequest},
		"malformed id":       {"DELETE", "/api/v1/saves/nope", "", http.StatusNotFound},
		"unknown id":         {"PATCH", "/api/v1/saves/" + uuid.NewString(), `{"name":"x"}`, http.StatusNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := call(t, saves, "alice", tc.method, tc.path, tc.body)
			if rec.Code != tc.status || rec.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("%d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestDeleteRefusesASaveThatAServerStillHolds(t *testing.T) {
	saves := &memorySaves{}
	record, _ := saves.CreateSave(context.Background(), "alice", "world", MaxSavesPerUser)
	saves.inUse = map[uuid.UUID]bool{record.ID: true}

	rec := call(t, saves, "alice", "DELETE", "/api/v1/saves/"+record.ID.String(), "")

	if rec.Code != http.StatusConflict || len(saves.records) != 1 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestDeleteKeepsTheSaveWhenTheBucketFails(t *testing.T) {
	saves := &memorySaves{purgeErr: errors.New("oracle: connection refused to 10.0.0.9")}
	record, _ := saves.CreateSave(context.Background(), "alice", "world", MaxSavesPerUser)

	rec := call(t, saves, "alice", "DELETE", "/api/v1/saves/"+record.ID.String(), "")

	if rec.Code != http.StatusInternalServerError || len(saves.records) != 1 || strings.Contains(rec.Body.String(), "10.0.0.9") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestExpireIdleDeletesIdleSavesAndSkipsOnesInUse(t *testing.T) {
	saves := &memorySaves{}
	idle, _ := saves.CreateSave(context.Background(), "alice", "idle", MaxSavesPerUser)
	held, _ := saves.CreateSave(context.Background(), "alice", "held", MaxSavesPerUser)
	saves.idle = []uuid.UUID{held.ID, idle.ID}
	saves.inUse = map[uuid.UUID]bool{held.ID: true}

	expired, err := NewService(saves, saves).ExpireIdle(context.Background())

	if err != nil || len(expired) != 1 || expired[0] != idle.ID.String() {
		t.Fatalf("expired = %v, err = %v", expired, err)
	}
	if len(saves.records) != 1 || saves.records[0].ID != held.ID {
		t.Fatalf("records = %+v", saves.records)
	}
}

func TestExpireIdleLeavesASaveThatWasUsedAfterItWasListed(t *testing.T) {
	saves := &memorySaves{}
	played, _ := saves.CreateSave(context.Background(), "alice", "played", MaxSavesPerUser)
	saves.idle = []uuid.UUID{played.ID}
	saves.revived = map[uuid.UUID]bool{played.ID: true}

	expired, err := NewService(saves, saves).ExpireIdle(context.Background())

	if err != nil || len(expired) != 0 || len(saves.records) != 1 {
		t.Fatalf("expired = %v, err = %v, records = %+v", expired, err, saves.records)
	}
}

func TestExpireIdleReportsBucketFailuresAndContinues(t *testing.T) {
	saves := &memorySaves{purgeErr: errors.New("bucket down")}
	first, _ := saves.CreateSave(context.Background(), "alice", "first", MaxSavesPerUser)
	saves.idle = []uuid.UUID{first.ID}

	expired, err := NewService(saves, saves).ExpireIdle(context.Background())

	if err == nil || len(expired) != 0 || len(saves.records) != 1 {
		t.Fatalf("expired = %v, err = %v", expired, err)
	}
}

func TestSavesExpireAfterNinetyIdleDays(t *testing.T) {
	saves := &memorySaves{}

	if _, err := NewService(saves, saves).ExpireIdle(context.Background()); err != nil {
		t.Fatal(err)
	}

	if saves.idleFor != 90*24*time.Hour {
		t.Fatalf("saves idle for %v are expired", saves.idleFor)
	}
}
