package customizers

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/configuration"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

const (
	defaultDifficulty = "normal"
)

var gameplayEnforcedOpts = map[definition.Option]string{
	definition.EULA: "true",
}

type gameplayCustomizer struct{}

func NewGameplayCustomizer() configuration.Customizer {
	return &gameplayCustomizer{}
}

func (gc *gameplayCustomizer) Customize(ctx context.Context, env map[definition.Option]string) (map[definition.Option]string, error) {
	for key, val := range gameplayEnforcedOpts {
		env[key] = val
	}

	if env[definition.DIFFICULTY] == "" {
		env[definition.DIFFICULTY] = defaultDifficulty
	}

	return env, nil
}
