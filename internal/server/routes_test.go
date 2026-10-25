package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/problem"
)

func TestProvisionFailure(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
	}{
		"invalid definition": {definition.InvalidDefinitionError{Message: "tier gold is not available"}, http.StatusUnprocessableEntity},
		"name in use":        {fmt.Errorf("wrapped: %w", ErrNameInUse), http.StatusConflict},
		"unexpected":         {errors.New("pq: password authentication failed for user admin"), http.StatusInternalServerError},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", nil)

			if err := provisionFailure(rec, req, tc.err); err != nil {
				t.Fatal(err)
			}

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if got := rec.Header().Get("Content-Type"); got != problem.ContentType {
				t.Fatalf("content type = %q", got)
			}
			if tc.status == http.StatusInternalServerError && strings.Contains(rec.Body.String(), "password") {
				t.Fatalf("internal message leaked: %s", rec.Body)
			}
		})
	}
}

func TestProvisionServerRejectsMalformedBody(t *testing.T) {
	h := NewServerHandlers(Service{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", strings.NewReader("{nope"))

	if err := h.provisionServer(rec, req); err != nil {
		t.Fatal(err)
	}

	var body problem.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest || body.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %+v", rec.Code, body)
	}
}

func TestProvisionServerRejectsTrailingData(t *testing.T) {
	h := NewServerHandlers(Service{})
	valid := `{"name":"my-server","durationMs":120000,"tier":"wooden"}`

	for _, suffix := range []string{" garbage", "}", "]", valid} {
		t.Run(suffix, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", strings.NewReader(valid+suffix))

			if err := h.provisionServer(rec, req); err != nil {
				t.Fatal(err)
			}

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", rec.Code)
			}
		})
	}
}
