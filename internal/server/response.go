package server

import (
	"encoding/json"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"time"
)

const operatorStatusUp = "up"

func newResponse(record store.ServerRecord, status server.Status, expiresAt time.Time) (server.Response, error) {
	var stored definition.ServerDefinition
	if err := json.Unmarshal(record.Definition, &stored); err != nil {
		return server.Response{}, fmt.Errorf("decode definition of server %s: %w", record.ID, err)
	}

	return server.Response{
		ID:        record.ID,
		Name:      record.Name,
		Hostname:  record.Hostname,
		Status:    status,
		Tier:      string(stored.Tier),
		Preset:    stored.Preset,
		Options:   definition.UserOptions(stored.Options),
		CreatedAt: record.CreatedAt,
		ExpiresAt: expiresAt,
	}, nil
}

// The cluster counts the duration from the moment the server started, which can be later than the stored expiry.
func observedResponse(record store.ServerRecord, resource *unstructured.Unstructured, now time.Time) (server.Response, error) {
	expiresAt := record.ExpiresAt
	status := server.StatusStopped

	if resource != nil {
		status = server.StatusProvisioning
		if operatorStatus, _, _ := unstructured.NestedString(resource.Object, "status", "status"); operatorStatus == operatorStatusUp {
			status = server.StatusRunning
		}
		if startedExpiry, ok := expiryFromStart(resource.Object); ok {
			expiresAt = startedExpiry
		}
	}

	if !expiresAt.After(now) {
		status = server.StatusExpired
	}

	return newResponse(record, status, expiresAt)
}

func expiryFromStart(resource map[string]interface{}) (time.Time, bool) {
	started, _, _ := unstructured.NestedString(resource, "status", "startedTime")
	startedAt, err := time.Parse(time.RFC3339, started)
	if err != nil {
		return time.Time{}, false
	}

	expireAfter, _, _ := unstructured.NestedInt64(resource, "spec", "expireAfter")
	return startedAt.Add(time.Duration(expireAfter) * time.Millisecond), true
}
