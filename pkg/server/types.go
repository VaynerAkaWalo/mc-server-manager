package server

import (
	"github.com/google/uuid"
	"time"
)

type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusRunning      Status = "running"
	StatusStopped      Status = "stopped"
	StatusExpired      Status = "expired"
)

type Response struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Hostname  string            `json:"hostname"`
	Status    Status            `json:"status"`
	Tier      string            `json:"tier"`
	Preset    string            `json:"preset,omitempty"`
	Options   map[string]string `json:"options"`
	CreatedAt time.Time         `json:"createdAt"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

type Request struct {
	Name       string            `json:"name"`
	DurationMs int64             `json:"durationMs"`
	Tier       string            `json:"tier"`
	Preset     string            `json:"preset,omitempty"`
	Options    map[string]string `json:"options"`
}

type Collection struct {
	Items      []Response `json:"items"`
	NextCursor string     `json:"nextCursor"`
}
