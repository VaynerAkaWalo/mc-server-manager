package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/google/uuid"
)

func TestCollectionQueryValidation(t *testing.T) {
	for _, raw := range []string{"status=unknown", "sort=name", "limit=0", "limit=-1", "limit=101", "limit=no", "cursor=no", "cursor=e30"} {
		t.Run(raw, func(t *testing.T) {
			h := NewServerHandlers(Service{})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/servers?"+raw, nil)
			if err := h.listServersHandler(rec, req); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusBadRequest || rec.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("response = %d %s", rec.Code, rec.Body)
			}
		})
	}
	query, err := parseCollectionQuery(url.Values{})
	if err != nil || query.Limit != 20 || query.Sort != "-createdAt" || query.Filter.Status != "active" {
		t.Fatalf("default = %+v, %v", query, err)
	}
}

func TestCollectionCursorRoundTrip(t *testing.T) {
	for _, sort := range []string{"createdAt", "-createdAt", "expiresAt", "-expiresAt"} {
		query, err := parseCollectionQuery(url.Values{"sort": {sort}, "status": {"expired"}, "limit": {"100"}})
		if err != nil {
			t.Fatal(err)
		}
		record := testRecord(time.Now().Add(-time.Hour))
		raw := encodeCursor(query, record)
		next, err := parseCollectionQuery(url.Values{"sort": {sort}, "status": {"expired"}, "cursor": {raw}})
		if err != nil {
			t.Fatal(err)
		}
		value := record.CreatedAt
		if sort == "expiresAt" || sort == "-expiresAt" {
			value = record.ExpiresAt
		}
		if next.After.ID != record.ID || !next.After.Value.Equal(value) {
			t.Fatalf("next = %+v", next)
		}
		if _, err := parseCollectionQuery(url.Values{"sort": {sort}, "status": {"running"}, "cursor": {raw}}); err == nil {
			t.Fatal("accepted mismatched filter")
		}
		if _, err := parseCollectionQuery(url.Values{"sort": {"-" + sort}, "status": {"expired"}, "cursor": {raw}}); err == nil {
			t.Fatal("accepted mismatched sort")
		}
	}
}

func TestCollectionResponse(t *testing.T) {
	first := testRecord(time.Now().Add(time.Hour))
	first.Status = "new"
	second := first
	second.ID = uuid.New()
	svc := CreateServerService(nil, &memoryStore{records: []store.ServerRecord{first, second}})
	query, _ := parseCollectionQuery(url.Values{"limit": {"1"}})
	page, err := svc.listServers(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != first.ID || page.NextCursor == "" {
		t.Fatalf("page = %+v", page)
	}

	svc = CreateServerService(nil, &memoryStore{})
	page, err = svc.listServers(context.Background(), query)
	if err != nil || page.Items == nil || page.NextCursor != "" {
		t.Fatalf("empty page = %+v, %v", page, err)
	}
}

func TestStatsRejectsInvalidStatus(t *testing.T) {
	h := NewServerHandlers(Service{})
	rec := httptest.NewRecorder()
	if err := h.statsHandler(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers/stats?status=no", nil)); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
