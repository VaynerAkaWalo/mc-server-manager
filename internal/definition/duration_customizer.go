package definition

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
	"time"
)

const MaxServerDuration = 14 * 24 * time.Hour

func customizeDuration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Duration.Minutes() < 2 {
		return definition, xhttp.NewError("duration cannot be less than 2 minutes", http.StatusBadRequest)
	}

	if MaxServerDuration < definition.Duration {
		return definition, xhttp.NewError("server cannot be provisioned for more than 14 days", http.StatusBadRequest)
	}

	return definition, nil
}
