package definition

import (
	"context"
	"time"
)

const MaxServerDuration = 14 * 24 * time.Hour

func customizeDuration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Duration.Minutes() < 2 {
		return definition, invalidf("duration cannot be less than 2 minutes")
	}

	if MaxServerDuration < definition.Duration {
		return definition, invalidf("server cannot be provisioned for more than 14 days")
	}

	return definition, nil
}
