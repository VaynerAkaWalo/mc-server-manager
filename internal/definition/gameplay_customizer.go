package definition

import "context"

const (
	defaultDifficulty = "normal"
)

var gameplayEnforcedOpts = map[Option]string{
	EULA: "true",
}

func customizeGameplay(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	for key, val := range gameplayEnforcedOpts {
		definition.Options[key] = val
	}

	if definition.Options[DIFFICULTY] == "" {
		definition.Options[DIFFICULTY] = defaultDifficulty
	}

	return definition, nil
}
