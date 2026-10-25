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
	case "active", "all", "new", "running", "stopped", "expired":
		return store.ServerFilter{Status: status}, nil
	default:
		return store.ServerFilter{}, errors.New("status must be active, all, new, running, stopped or expired")
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
		cursor := collectionCursor{}
		body, err := base64.RawURLEncoding.DecodeString(raw)
		if len(raw) > 4096 || err != nil || json.Unmarshal(body, &cursor) != nil || cursor.Version != 1 || cursor.Boundary.Value.IsZero() || cursor.Boundary.ID == uuid.Nil || cursor.Status != query.Filter.Status || cursor.Sort != query.Sort {
			return store.ServerQuery{}, errors.New("cursor is invalid or does not match the status and sort")
		}
		query.After = &cursor.Boundary
	}
	return query, nil
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
