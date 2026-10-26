package definition

import (
	"context"
	"regexp"
	"strings"
)

// Game servers share the minecraft-server namespace and are named after the server, so these would take over the stack's own resources.
var reservedNames = map[string]bool{
	"mc-proxy":           true,
	"mc-server-manager":  true,
	"mc-server-operator": true,
}

func customizeName(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	definition.Name = strings.ToLower(strings.ReplaceAll(definition.Name, " ", "-"))

	pattern := regexp.MustCompile("[^a-zA-Z0-9-]+")

	sanitizedString := pattern.ReplaceAllString(definition.Name, "")

	if len(sanitizedString) < 5 || len(sanitizedString) > 25 {
		return definition, invalidf("the server name must be between 5 and 25 characters long")
	}

	if reservedNames[strings.ToLower(definition.Name)] {
		return definition, invalidf("the server name %s is reserved", definition.Name)
	}

	return definition, nil
}
