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

func newResponse(record store.ServerRecord, status server.Status) (server.Response, error) {
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
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func observedResponse(record store.ServerRecord, resource *unstructured.Unstructured, now time.Time) (server.Response, error) {
	status := server.StatusStopped

	if resource != nil {
		status = server.StatusProvisioning
		if operatorStatus, _, _ := unstructured.NestedString(resource.Object, "status", "status"); operatorStatus == operatorStatusUp {
			status = server.StatusRunning
		}
	}

	if !record.ExpiresAt.After(now) {
		status = server.StatusExpired
	}

	return newResponse(record, status)
}
