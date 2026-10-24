package definition

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type optionValidator func(value string) bool

var freeText = regexp.MustCompile(`^[\p{L}\p{N} _.,:'!?-]*$`).MatchString
var identifierList = regexp.MustCompile(`^[a-zA-Z0-9_.:, -]*$`).MatchString

var serverTypes = []string{"VANILLA", "PAPER", "FABRIC", "FORGE", "NEOFORGE", "QUILT"}

// Mods outside Modrinth are arbitrary jars, so only ones reviewed for the frontend templates are allowed.
var approvedModUrls = []string{
	"https://github.com/marin774/lockout-fabric/releases/download/v0.9.3/lockout-fabric-0.9.3.jar",
}

var userSettableOpts = map[Option]optionValidator{
	VERSION:           identifierList,
	ONLINE_MODE:       identifierList,
	TYPE:              isServerType,
	MOTD:              freeText,
	OPS:               identifierList,
	MAX_PLAYERS:       identifierList,
	DIFFICULTY:        identifierList,
	MODRINTH_PROJECTS: identifierList,
	MODS:              areApprovedMods,
	VIEW_DISTANCE:     identifierList,
}

func customizeOptions(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Options == nil {
		definition.Options = make(map[Option]string)
	}

	for key, value := range definition.Options {
		isValid, settable := userSettableOpts[key]
		if !settable {
			return definition, fmt.Errorf("option %s cannot be set", key)
		}

		if !isValid(value) {
			return definition, fmt.Errorf("option %s has an unsupported value", key)
		}
	}

	return definition, nil
}

func isServerType(value string) bool {
	return slices.Contains(serverTypes, value)
}

func areApprovedMods(value string) bool {
	for _, url := range strings.Split(value, ",") {
		if !slices.Contains(approvedModUrls, strings.TrimSpace(url)) {
			return false
		}
	}

	return true
}
