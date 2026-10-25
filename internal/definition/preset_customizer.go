package definition

import (
	"context"
)

// Mods are arbitrary code, so they only come from presets reviewed here and never from user options.
var presets = map[string]map[Option]string{
	"lockout": {
		VERSION:           "1.21.3",
		TYPE:              "FABRIC",
		MODRINTH_PROJECTS: "fabric-api,lithium",
		MODS:              "https://github.com/marin774/lockout-fabric/releases/download/v0.9.3/lockout-fabric-0.9.3.jar",
	},
}

func customizePreset(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Preset == "" {
		return definition, nil
	}

	presetOpts, found := presets[definition.Preset]
	if !found {
		return definition, invalidf("preset %s does not exist", definition.Preset)
	}

	for key, value := range presetOpts {
		definition.Options[key] = value
	}

	return definition, nil
}
