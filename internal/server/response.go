package server

import (
	"encoding/json"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
)

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
