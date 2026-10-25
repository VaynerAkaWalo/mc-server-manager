package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type ServerRecord struct {
	ID         uuid.UUID
	Name       string
	Hostname   string
	Definition json.RawMessage
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type NewServer struct {
	Name       string
	Hostname   string
	Definition any
	ExpiresAt  time.Time
}

func (s *Store) CreateServer(ctx context.Context, server NewServer) (ServerRecord, error) {
	definition, err := json.Marshal(server.Definition)
	if err != nil {
		return ServerRecord{}, fmt.Errorf("marshal definition: %w", err)
	}

	record := ServerRecord{
		Name:       server.Name,
		Hostname:   server.Hostname,
		Definition: definition,
		ExpiresAt:  server.ExpiresAt,
	}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO mc_servers (name, hostname, definition, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		record.Name, record.Hostname, record.Definition, record.ExpiresAt,
	).Scan(&record.ID, &record.CreatedAt)
	if err != nil {
		return ServerRecord{}, fmt.Errorf("insert server: %w", err)
	}

	return record, nil
}

func (s *Store) DeleteServer(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM mc_servers WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}

	return nil
}
