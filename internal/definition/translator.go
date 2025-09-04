package definition

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
)

type customizer func(ctx context.Context, definition ServerDefinition) (ServerDefinition, error)

func TranslateDefinition(ctx context.Context, definition ServerDefinition) (*v1alpha1.McServerSpec, error) {
	customizers := []customizer{
		customizeName,
		customizeDuration,
		customizeOperators,
		customizeContainer,
		customizeGameplay,
		customizeWorldGeneration,
	}

	var err error
	for _, definitionCustomizer := range customizers {
		definition, err = definitionCustomizer(ctx, definition)
		if err != nil {
			return nil, err
		}
	}

	instance := translateTier(ctx, definition.Tier)

	return &v1alpha1.McServerSpec{
		Name:         definition.Name,
		Image:        defaultImage,
		Env:          stringifyEnvironment(ctx, definition.Options),
		Memory:       instance.memory,
		InstanceType: instance.instanceType,
		ExpireAfter:  definition.Duration.Milliseconds(),
	}, nil
}

func stringifyEnvironment(ctx context.Context, environment map[Option]string) map[string]string {
	postProcessed := make(map[string]string, len(environment))

	for key, value := range environment {
		stringKey := optsKeys[key]
		postProcessed[stringKey] = value
	}

	return postProcessed
}
