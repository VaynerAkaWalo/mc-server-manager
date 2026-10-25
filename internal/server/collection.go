package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/google/uuid"
)

type collectionCursor struct {
	Version  int                  `json:"v"`
	Status   string               `json:"status"`
	Sort     string               `json:"sort"`
	Boundary store.ServerBoundary `json:"boundary"`
}

func parseFilter(values url.Values) (store.ServerFilter, error) {
	status := values.Get("status")
	if status == "" {
		status = "active"
	}
	switch status {
	case "active", "all", "new", "running", "degraded", "expiring", "stopped", "expired":
		return store.ServerFilter{Status: status}, nil
	default:
		return store.ServerFilter{}, errors.New("status must be active, all, new, running, degraded, expiring, stopped or expired")
	}
}

func parseCollectionQuery(values url.Values) (store.ServerQuery, error) {
	filter, err := parseFilter(values)
	if err != nil {
		return store.ServerQuery{}, err
	}
	query := store.ServerQuery{Filter: filter, Sort: values.Get("sort"), Limit: 20}
	if query.Sort == "" {
		query.Sort = "-createdAt"
	}
	switch query.Sort {
	case "createdAt", "-createdAt", "expiresAt", "-expiresAt":
	default:
		return store.ServerQuery{}, errors.New("sort must be createdAt, -createdAt, expiresAt or -expiresAt")
	}
	if raw := values.Get("limit"); raw != "" {
		query.Limit, err = strconv.Atoi(raw)
		if err != nil || query.Limit < 1 || query.Limit > 100 {
			return store.ServerQuery{}, errors.New("limit must be between 1 and 100")
		}
	}
	if raw := values.Get("cursor"); raw != "" {
		cursor, err := decodeCursor(raw, query)
		if err != nil {
			return store.ServerQuery{}, err
		}
		query.After = &cursor.Boundary
	}
	return query, nil
}

func decodeCursor(raw string, query store.ServerQuery) (collectionCursor, error) {
	var cursor collectionCursor
	invalid := errors.New("cursor is invalid or does not match the status and sort")
	if len(raw) > 4096 {
		return cursor, invalid
	}

	body, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || json.Unmarshal(body, &cursor) != nil {
		return cursor, invalid
	}
	if !cursor.matchesQuery(query) {
		return cursor, invalid
	}
	return cursor, nil
}

func (c collectionCursor) matchesQuery(query store.ServerQuery) bool {
	if c.Version != 1 || c.Boundary.Value.IsZero() || c.Boundary.ID == uuid.Nil {
		return false
	}
	return c.Status == query.Filter.Status && c.Sort == query.Sort
}

func encodeCursor(query store.ServerQuery, last store.ServerRecord) string {
	value := last.CreatedAt
	if strings.TrimPrefix(query.Sort, "-") == "expiresAt" {
		value = last.ExpiresAt
	}
	cursor := collectionCursor{Version: 1, Status: query.Filter.Status, Sort: query.Sort, Boundary: store.ServerBoundary{Value: value, ID: last.ID}}
	body, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(body)
}
