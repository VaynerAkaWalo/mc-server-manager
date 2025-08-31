package customizers

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/configuration"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

var containerEnforcedOpts = map[definition.Option]string{
	definition.MEMORY:          "",
	definition.USE_AIKAR_FLAGS: "true",
	definition.JVM_XX_OPTS:     "-XX:MaxRAMPercentage=75",
}

type containerCustomizer struct {
}

func NewContainerCustomizer() configuration.Customizer {
	return &containerCustomizer{}
}

func (cc *containerCustomizer) Customize(ctx context.Context, env map[definition.Option]string) (map[definition.Option]string, error) {
	for key, val := range containerEnforcedOpts {
		env[key] = val
	}

	return env, nil
}
