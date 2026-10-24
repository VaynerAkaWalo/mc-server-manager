package definition

import "time"

type (
	ServerDefinition struct {
		Name     string            `json:"name"`
		Options  map[Option]string `json:"options"`
		Tier     Tier              `json:"tier"`
		Preset   string            `json:"preset,omitempty"`
		Duration time.Duration     `json:"duration"`
		SaveId   string            `json:"saveId,omitempty"`
	}
)
