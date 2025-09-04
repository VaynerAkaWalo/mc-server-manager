package definition

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func customizeName(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	definition.Name = strings.ReplaceAll(definition.Name, " ", "-")

	pattern := regexp.MustCompile("[^a-zA-Z0-9-]+")

	sanitizedString := pattern.ReplaceAllString(definition.Name, "")

	if len(sanitizedString) < 5 || len(sanitizedString) > 25 {
		return definition, fmt.Errorf("the server name must be between 5 and 25 characters long")
	}

	return definition, nil
}
