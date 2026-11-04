package internalapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type fakeExpirer struct {
	names []string
	err   error
}

func (f fakeExpirer) ShutdownExpired() ([]string, error) { return f.names, f.err }

type fakeSaveExpirer struct {
	ids []string
	err error
}

func (f fakeSaveExpirer) ExpireIdle(context.Context) ([]string, error) { return f.ids, f.err }

func serve(h *Handlers, method, path string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	h.RegisterRoutes(&xhttp.Router{ServeMux: mux})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestReady(t *testing.T) {
	up := func(context.Context) error { return nil }
	down := func(context.Context) error { return errors.New("connection refused to 10.0.0.5") }

	if rec := serve(NewHandlers(up, fakeExpirer{}, nil), "GET", "/ready"); rec.Code != http.StatusOK {
		t.Fatalf("up: status = %d", rec.Code)
	}

	rec := serve(NewHandlers(down, fakeExpirer{}, nil), "GET", "/ready")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("down: status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Fatalf("internal error leaked: %s", rec.Body)
	}
}

func TestExpireServers(t *testing.T) {
	h := NewHandlers(nil, fakeExpirer{names: []string{"old-server"}}, nil)
	rec := serve(h, "POST", "/internal/v1/jobs/expire-servers")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"expired":["old-server"]`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	h = NewHandlers(nil, fakeExpirer{err: errors.New("k8s boom")}, nil)
	rec = serve(h, "POST", "/internal/v1/jobs/expire-servers")
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestExpireSaves(t *testing.T) {
	h := NewHandlers(nil, fakeExpirer{}, fakeSaveExpirer{ids: []string{"4f9d0a3e-7c1b-4d6f-9a52-0e1c8b7d2a10"}})
	rec := serve(h, "POST", "/internal/v1/jobs/expire-saves")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"expired":["4f9d0a3e-7c1b-4d6f-9a52-0e1c8b7d2a10"]`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	h = NewHandlers(nil, fakeExpirer{}, fakeSaveExpirer{err: errors.New("bucket boom")})
	rec = serve(h, "POST", "/internal/v1/jobs/expire-saves")
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
