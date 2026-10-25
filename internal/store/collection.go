package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ServerFilter struct {
	Status string
}

type ServerQuery struct {
	Filter ServerFilter
	Sort   string
	Limit  int
	After  *ServerBoundary
}

type ServerBoundary struct {
	Value time.Time
	ID    uuid.UUID
}

const serverFilterSQL = `($1 = 'all'
 OR ($1 = 'active' AND status IN ('new', 'running', 'degraded', 'expiring') AND expires_at > now())
 OR status = $1)`

func (s *Store) CountServers(ctx context.Context, filter ServerFilter) (ServerCounts, error) {
	var counts ServerCounts
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE status IN ('new', 'running', 'degraded', 'expiring') AND expires_at > now()), count(*)
   FROM mc_servers WHERE `+serverFilterSQL, filter.Status,
	).Scan(&counts.Online, &counts.Created)
	if err != nil {
		return ServerCounts{}, fmt.Errorf("count servers: %w", err)
	}
	return counts, nil
}

func (s *Store) ListServers(ctx context.Context, query ServerQuery) ([]ServerRecord, error) {
	column := "created_at"
	if strings.TrimPrefix(query.Sort, "-") == "expiresAt" {
		column = "expires_at"
	}
	direction, comparison := "ASC", ">"
	if strings.HasPrefix(query.Sort, "-") {
		direction, comparison = "DESC", "<"
	}

	var afterValue any
	var afterID any
	if query.After != nil {
		afterValue, afterID = query.After.Value, query.After.ID
	}
	statement := fmt.Sprintf(`SELECT id, name, hostname, definition, created_at, expires_at, last_active_at, status
  FROM mc_servers
  WHERE %s AND ($2::timestamptz IS NULL OR (%s, id) %s ($2, $3::uuid))
  ORDER BY %s %s, id %s LIMIT $4`, serverFilterSQL, column, comparison, column, direction, direction)

	rows, err := s.pool.Query(ctx, statement, query.Filter.Status, afterValue, afterID, query.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Hostname, &record.Definition, &record.CreatedAt, &record.ExpiresAt, &record.LastActiveAt, &record.Status); err != nil {
			return nil, fmt.Errorf("scan server: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	return records, nil
}
