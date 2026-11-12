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

	if rec := serve(NewHandlers(up), "GET", "/ready"); rec.Code != http.StatusOK {
		t.Fatalf("up: status = %d", rec.Code)
	}

	rec := serve(NewHandlers(down), "GET", "/ready")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("down: status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Fatalf("internal error leaked: %s", rec.Body)
	}
}
