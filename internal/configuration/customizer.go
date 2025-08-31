package configuration

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/configuration/customizers"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

type Customizer interface {
	Customize(context.Context, map[definition.Option]string) (map[definition.Option]string, error)
}

type combinedCustomizer struct {
	customizers []Customizer
}

func BuildCustomizer(customizers ...Customizer) Customizer {
	return &combinedCustomizer{
		customizers: customizers,
	}
}

func NewDefaultEnvCustomizer() Customizer {
	return BuildCustomizer(
		customizers.NewContainerCustomizer(),
		customizers.NewOperatorsCustomizer(),
		customizers.NewWorldGeneratorCustomizer(),
		customizers.NewGameplayCustomizer())
}

func (cc *combinedCustomizer) Customize(ctx context.Context, env map[definition.Option]string) (map[definition.Option]string, error) {
	currentEnv := env
	var err error

	for _, processor := range cc.customizers {
		currentEnv, err = processor.Customize(ctx, currentEnv)

		if err != nil {
			return currentEnv, err
		}
	}

	return currentEnv, nil
}
