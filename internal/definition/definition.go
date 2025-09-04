package definition

import "time"

type (
	ServerDefinition struct {
		Name     string            `json:"name"`
		Options  map[Option]string `json:"options"`
		Tier     Tier              `json:"tier"`
		Duration time.Duration     `json:"duration"`
	}
)
