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
	AsOf   time.Time
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

func serverPredicate(filter ServerFilter) (string, []any) {
	active := "status IN ('provisioning', 'running') AND expires_at > $1"
	predicate := active
	switch filter.Status {
	case "all":
		predicate = "TRUE"
	case "provisioning", "running":
		predicate = "status = $2 AND expires_at > $1"
	case "stopped":
		predicate = "status = $2"
	case "expired":
		predicate = "(status = 'expired' OR (status IN ('provisioning', 'running') AND expires_at <= $1))"
	}

	args := []any{filter.AsOf}
	if strings.Contains(predicate, "$2") {
		args = append(args, filter.Status)
	}
	return "created_at <= $1 AND (" + predicate + ")", args
}

func (s *Store) CountServers(ctx context.Context, filter ServerFilter) (ServerCounts, error) {
	predicate, args := serverPredicate(filter)
	var counts ServerCounts
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE status IN ('provisioning', 'running') AND expires_at > $1), count(*)
   FROM mc_servers WHERE `+predicate, args...,
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

	predicate, args := serverPredicate(query.Filter)
	if query.After != nil {
		predicate += fmt.Sprintf(" AND (%s, id) %s ($%d, $%d)", column, comparison, len(args)+1, len(args)+2)
		args = append(args, query.After.Value, query.After.ID)
	}
	args = append(args, query.Limit+1)
	statement := fmt.Sprintf(`SELECT id, name, hostname, definition, created_at, expires_at,
  CASE WHEN status IN ('provisioning', 'running') AND expires_at <= $1 THEN 'expired' ELSE status END
  FROM mc_servers WHERE %s ORDER BY %s %s, id %s LIMIT $%d`, predicate, column, direction, direction, len(args))

	rows, err := s.pool.Query(ctx, statement, args...)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Hostname, &record.Definition, &record.CreatedAt, &record.ExpiresAt, &record.Status); err != nil {
			return nil, fmt.Errorf("scan server: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	return records, nil
}
