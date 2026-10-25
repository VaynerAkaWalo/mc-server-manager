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

type ServerCounts struct {
	Online  int
	Created int
}

func (s *Store) CountServers(ctx context.Context) (ServerCounts, error) {
	var counts ServerCounts
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE expires_at > now()), count(*) FROM mc_servers`,
	).Scan(&counts.Online, &counts.Created)
	if err != nil {
		return ServerCounts{}, fmt.Errorf("count servers: %w", err)
	}

	return counts, nil
}

// A late cluster start can keep the latest run alive past its stored expiry.
func (s *Store) ListServers(ctx context.Context, clusterNames []string) ([]ServerRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, hostname, definition, created_at, expires_at
		 FROM mc_servers
		 WHERE expires_at > now()
		    OR id IN (SELECT DISTINCT ON (name) id FROM mc_servers WHERE name = ANY($1) ORDER BY name, created_at DESC, id DESC)
		 ORDER BY created_at DESC, id DESC`,
		clusterNames,
	)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Hostname, &record.Definition, &record.CreatedAt, &record.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan server: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}

	return records, nil
}
