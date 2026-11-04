package definition

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
)

type customizer func(ctx context.Context, definition ServerDefinition) (ServerDefinition, error)

func TranslateDefinition(ctx context.Context, definition ServerDefinition, versions VersionSource) (*v1alpha1.McServerSpec, error) {
	customizers := []customizer{
		customizeOptions(versions),
		customizePreset,
		customizeVersionRequired,
		customizePerformance(versions),
		customizeName,
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

	instance, err := translateTier(ctx, definition.Tier)
	if err != nil {
		return nil, err
	}

	spec := &v1alpha1.McServerSpec{
		Name:         definition.Name,
		Image:        defaultImage,
		Env:          stringifyEnvironment(ctx, definition.Options),
		Memory:       instance.memory,
		InstanceType: instance.instanceType,
		ExpireAfter:  definition.Duration.Milliseconds(),
	}
	if definition.SaveId != "" {
		spec.Save = &v1alpha1.SaveSpec{ID: definition.SaveId}
	}

	return spec, nil
}

func stringifyEnvironment(ctx context.Context, environment map[Option]string) map[string]string {
	postProcessed := make(map[string]string, len(environment))

	for key, value := range environment {
		stringKey := optsKeys[key]
		postProcessed[stringKey] = value
	}

	return postProcessed
}
