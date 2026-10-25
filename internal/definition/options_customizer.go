package definition

import (
	"context"
	"regexp"
	"slices"
)

type optionSpec struct {
	valid  func(value string) bool
	values []string
}

var freeText = regexp.MustCompile(`^[\p{L}\p{N} _.,:'!?-]*$`).MatchString
var identifierList = regexp.MustCompile(`^[a-zA-Z0-9_.:, -]*$`).MatchString

func oneOf(values ...string) optionSpec {
	return optionSpec{valid: func(value string) bool { return slices.Contains(values, value) }, values: values}
}

func text(valid func(string) bool) optionSpec {
	return optionSpec{valid: valid}
}

var userSettableOpts = map[Option]optionSpec{
	VERSION:       text(identifierList),
	ONLINE_MODE:   oneOf("true", "false"),
	TYPE:          oneOf("VANILLA", "PAPER", "FABRIC", "FORGE", "NEOFORGE", "QUILT"),
	MOTD:          text(freeText),
	OPS:           text(identifierList),
	MAX_PLAYERS:   text(identifierList),
	DIFFICULTY:    oneOf("peaceful", "easy", "normal", "hard"),
	VIEW_DISTANCE: text(identifierList),
}

func customizeOptions(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Options == nil {
		definition.Options = make(map[Option]string)
	}

	for key, value := range definition.Options {
		spec, settable := userSettableOpts[key]
		if !settable {
			return definition, invalidf("option %s cannot be set", key)
		}

		if !spec.valid(value) {
			return definition, invalidf("option %s has an unsupported value", key)
		}
	}

	return definition, nil
}

func UserOptions(options map[Option]string) map[string]string {
	visible := make(map[string]string, len(options))
	for key, value := range options {
		if _, settable := userSettableOpts[key]; settable {
			visible[string(key)] = value
		}
	}

	return visible
}
