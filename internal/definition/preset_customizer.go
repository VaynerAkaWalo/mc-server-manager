package definition

import (
	"context"
)

// Mods are arbitrary code, so they only come from presets reviewed here and never from user options.
type preset struct {
	name        string
	description string
	options     map[Option]string
}

var presets = map[string]preset{
	"lockout": {
		name:        "Lockout",
		description: "Compete in bingo challenges to claim victory! This gameplay focuses on team-based objective completion.",
		options: map[Option]string{
			VERSION:           "1.21.3",
			TYPE:              "FABRIC",
			MODRINTH_PROJECTS: "fabric-api,lithium",
			MODS:              "https://github.com/marin774/lockout-fabric/releases/download/v0.9.3/lockout-fabric-0.9.3.jar",
		},
	},
}

func customizePreset(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Preset == "" {
		return definition, nil
	}

	selected, found := presets[definition.Preset]
	if !found {
		return definition, invalidf("preset %s does not exist", definition.Preset)
	}

	for key, value := range selected.options {
		definition.Options[key] = value
	}

	return definition, nil
}
