package definition

import (
	"context"
	"regexp"
	"slices"
)

type optionValidator func(value string) bool

var freeText = regexp.MustCompile(`^[\p{L}\p{N} _.,:'!?-]*$`).MatchString
var identifierList = regexp.MustCompile(`^[a-zA-Z0-9_.:, -]*$`).MatchString

var serverTypes = []string{"VANILLA", "PAPER", "FABRIC", "FORGE", "NEOFORGE", "QUILT"}

var userSettableOpts = map[Option]optionValidator{
	VERSION:       identifierList,
	ONLINE_MODE:   identifierList,
	TYPE:          isServerType,
	MOTD:          freeText,
	OPS:           identifierList,
	MAX_PLAYERS:   identifierList,
	DIFFICULTY:    identifierList,
	VIEW_DISTANCE: identifierList,
}

func customizeOptions(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Options == nil {
		definition.Options = make(map[Option]string)
	}

	for key, value := range definition.Options {
		isValid, settable := userSettableOpts[key]
		if !settable {
			return definition, invalidf("option %s cannot be set", key)
		}

		if !isValid(value) {
			return definition, invalidf("option %s has an unsupported value", key)
		}
	}

	return definition, nil
}

func isServerType(value string) bool {
	return slices.Contains(serverTypes, value)
}
