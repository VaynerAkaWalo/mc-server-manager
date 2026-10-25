package server

import (
	"github.com/google/uuid"
	"time"
)

type Status string

const (
	StatusNew      Status = "new"
	StatusRunning  Status = "running"
	StatusDegraded Status = "degraded"
	StatusStopped  Status = "stopped"
	StatusExpired  Status = "expired"
)

type Players struct {
	Online int `json:"online"`
	Max    int `json:"max"`
}

type Response struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Hostname       string            `json:"hostname"`
	Status         Status            `json:"status"`
	Players        *Players          `json:"players"`
	Tier           string            `json:"tier"`
	Preset         string            `json:"preset,omitempty"`
	Options        map[string]string `json:"options"`
	CreatedAt      time.Time         `json:"createdAt"`
	ExpiresAt      time.Time         `json:"expiresAt"`
	IdleShutdownAt *time.Time        `json:"idleShutdownAt"`
}

type Request struct {
	Name    string            `json:"name"`
	Tier    string            `json:"tier"`
	Preset  string            `json:"preset,omitempty"`
	Options map[string]string `json:"options"`
}

type Collection struct {
	Items      []Response `json:"items"`
	NextCursor string     `json:"nextCursor"`
}
