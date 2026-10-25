package server

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
)

func testRecord(expiresAt time.Time) store.ServerRecord {
	return store.ServerRecord{
		ID:         uuid.New(),
		Name:       "my-server",
		Hostname:   "my-server.mc-server.blamedevs.com",
		Definition: json.RawMessage(`{"tier":"wooden","preset":"lockout","options":{"MOTD":"hi","OPS":"abc","MEMORY":"","MODS":"https://x"}}`),
		CreatedAt:  expiresAt.Add(-time.Hour),
		ExpiresAt:  expiresAt,
	}
}

func TestResponseUsesStoredExpiry(t *testing.T) {
	record := testRecord(time.Now().Add(time.Minute))
	response, err := newResponse(record, server.StatusProvisioning)
	if err != nil {
		t.Fatal(err)
	}

	if !response.ExpiresAt.Equal(record.ExpiresAt) {
		t.Fatalf("expiresAt = %s, want stored %s", response.ExpiresAt, record.ExpiresAt)
	}
}

func TestResponseShape(t *testing.T) {
	response, err := newResponse(testRecord(time.Now().Add(time.Hour)), server.StatusProvisioning)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(response)
	for _, field := range []string{`"id"`, `"hostname"`, `"createdAt"`, `"expiresAt"`, `"tier":"wooden"`, `"preset":"lockout"`, `"MOTD":"hi"`} {
		if !strings.Contains(string(body), field) {
			t.Fatalf("missing %s in %s", field, body)
		}
	}
	for _, hidden := range []string{"MEMORY", "MODS", "IP", "remainingTime"} {
		if strings.Contains(string(body), hidden) {
			t.Fatalf("%s must not be exposed: %s", hidden, body)
		}
	}
}
