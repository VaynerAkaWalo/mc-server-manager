package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

const uniqueViolation = "23505"

var (
	ErrNotFound        = errors.New("server not found")
	ErrActiveNameTaken = errors.New("an active server already uses this name")
)

type ServerRecord struct {
	Status       string
	ID           uuid.UUID
	Name         string
	Hostname     string
	Definition   json.RawMessage
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastActiveAt *time.Time
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
		 RETURNING id, created_at, status`,
		record.Name, record.Hostname, record.Definition, record.ExpiresAt,
	).Scan(&record.ID, &record.CreatedAt, &record.Status)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return ServerRecord{}, ErrActiveNameTaken
	}
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

func (s *Store) GetServer(ctx context.Context, id uuid.UUID) (ServerRecord, error) {
	var record ServerRecord
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, hostname, definition, created_at, expires_at, last_active_at, status FROM mc_servers WHERE id = $1`, id,
	).Scan(&record.ID, &record.Name, &record.Hostname, &record.Definition, &record.CreatedAt, &record.ExpiresAt, &record.LastActiveAt, &record.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ServerRecord{}, ErrNotFound
	}
	if err != nil {
		return ServerRecord{}, fmt.Errorf("get server: %w", err)
	}

	return record, nil
}

func (s *Store) ListMonitoredServers(ctx context.Context) ([]ServerRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, status FROM mc_servers WHERE status IN ('new', 'running', 'degraded') AND expires_at > now()`,
	)
	if err != nil {
		return nil, fmt.Errorf("list monitored servers: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Status); err != nil {
			return nil, fmt.Errorf("scan monitored server: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list monitored servers: %w", err)
	}

	return records, nil
}

func (s *Store) MarkRunning(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mc_servers SET status = 'running', last_active_at = COALESCE(last_active_at, now())
		 WHERE id = $1 AND status IN ('new', 'degraded')`, id,
	); err != nil {
		return fmt.Errorf("mark server running: %w", err)
	}

	return nil
}

func (s *Store) MarkActive(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET last_active_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("mark server active: %w", err)
	}

	return nil
}

func (s *Store) ListIdleServers(ctx context.Context, idleFor time.Duration) ([]ServerRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, status FROM mc_servers
		 WHERE status IN ('running', 'degraded') AND last_active_at <= now() - make_interval(secs => $1)`,
		idleFor.Seconds(),
	)
	if err != nil {
		return nil, fmt.Errorf("list idle servers: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Status); err != nil {
			return nil, fmt.Errorf("scan idle server: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list idle servers: %w", err)
	}

	return records, nil
}

func (s *Store) ExpireServer(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mc_servers SET status = 'expired' WHERE id = $1 AND status IN ('new', 'running', 'degraded')`, id,
	); err != nil {
		return fmt.Errorf("expire server: %w", err)
	}

	return nil
}

func (s *Store) MarkDegraded(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mc_servers SET status = 'degraded' WHERE id = $1 AND status = 'running'`, id,
	); err != nil {
		return fmt.Errorf("mark server degraded: %w", err)
	}

	return nil
}

func (s *Store) MarkStopped(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mc_servers SET status = 'stopped' WHERE id = $1 AND status IN ('new', 'running', 'degraded')`, id,
	); err != nil {
		return fmt.Errorf("mark server stopped: %w", err)
	}

	return nil
}

func (s *Store) MarkExpired(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mc_servers SET status = 'expired' WHERE status IN ('new', 'running', 'degraded') AND expires_at <= now()`,
	); err != nil {
		return fmt.Errorf("mark servers expired: %w", err)
	}

	return nil
}
