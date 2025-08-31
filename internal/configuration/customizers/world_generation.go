package customizers

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/configuration"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"strconv"
)

const DefaultViewDistance = "10"

type worldGeneratorCustomizer struct {
}

func NewWorldGeneratorCustomizer() configuration.Customizer {
	return &worldGeneratorCustomizer{}
}

func (wgc *worldGeneratorCustomizer) Customize(ctx context.Context, env map[definition.Option]string) (map[definition.Option]string, error) {
	val, found := env[definition.VIEW_DISTANCE]
	if !found {
		env[definition.VIEW_DISTANCE] = DefaultViewDistance
	}

	viewDistance, err := strconv.Atoi(val)
	if err == nil || viewDistance < 1 || viewDistance > 64 {
		env[definition.VIEW_DISTANCE] = DefaultViewDistance
	}

	return env, nil
}
