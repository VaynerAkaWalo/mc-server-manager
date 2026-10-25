package server

import (
	"encoding/json"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"time"
)

func newResponse(record store.ServerRecord, status server.Status) (server.Response, error) {
	var stored definition.ServerDefinition
	if err := json.Unmarshal(record.Definition, &stored); err != nil {
		return server.Response{}, fmt.Errorf("decode definition of server %s: %w", record.ID, err)
	}

	var idleShutdownAt *time.Time
	if record.LastActiveAt != nil {
		shutdownAt := record.LastActiveAt.Add(idleTimeouts.Expired)
		idleShutdownAt = &shutdownAt
	}

	return server.Response{
		ID:             record.ID,
		Name:           record.Name,
		Hostname:       record.Hostname,
		Status:         status,
		Tier:           string(stored.Tier),
		Preset:         stored.Preset,
		Options:        definition.UserOptions(stored.Options),
		CreatedAt:      record.CreatedAt,
		ExpiresAt:      record.ExpiresAt,
		IdleShutdownAt: idleShutdownAt,
	}, nil
}
