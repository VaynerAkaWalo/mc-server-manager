package definition

import (
	"context"
	"time"
)

const (
	MinServerDuration = 2 * time.Minute
	MaxServerDuration = 14 * 24 * time.Hour
)

func customizeDuration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Duration < MinServerDuration {
		return definition, invalidf("duration cannot be less than 2 minutes")
	}

	if MaxServerDuration < definition.Duration {
		return definition, invalidf("server cannot be provisioned for more than 14 days")
	}

	return definition, nil
}
