package server

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func resource(operatorStatus, startedTime string, expireAfter time.Duration) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"spec":   map[string]any{"expireAfter": expireAfter.Milliseconds()},
		"status": map[string]any{"status": operatorStatus, "startedTime": startedTime},
	}}
}

func TestObservedResponseStatus(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	started := now.Add(-time.Minute).UTC().Format(time.RFC3339)

	cases := map[string]struct {
		expiresAt time.Time
		resource  *unstructured.Unstructured
		want      server.Status
	}{
		"new resource":            {future, resource("new", "", time.Hour), server.StatusProvisioning},
		"not ready":               {future, resource("not ready", "", time.Hour), server.StatusProvisioning},
		"up":                      {future, resource("up", started, time.Hour), server.StatusRunning},
		"no resource":             {future, nil, server.StatusStopped},
		"no resource and expired": {now.Add(-time.Hour), nil, server.StatusExpired},
		"still up after expiry":   {now.Add(-time.Hour), resource("up", now.Add(-2*time.Hour).UTC().Format(time.RFC3339), time.Hour), server.StatusExpired},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := observedResponse(testRecord(tc.expiresAt), tc.resource, now)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != tc.want {
				t.Fatalf("status = %s, want %s", response.Status, tc.want)
			}
		})
	}
}

func TestObservedResponseExpiryFollowsStart(t *testing.T) {
	now := time.Now()
	started := now.Add(-time.Minute).Truncate(time.Second)

	response, err := observedResponse(testRecord(now.Add(time.Minute)), resource("up", started.UTC().Format(time.RFC3339), time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}

	if !response.ExpiresAt.Equal(started.Add(time.Hour)) {
		t.Fatalf("expiresAt = %s", response.ExpiresAt)
	}
}

func TestResponseShape(t *testing.T) {
	response, err := newResponse(testRecord(time.Now().Add(time.Hour)), server.StatusProvisioning, time.Now())
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
