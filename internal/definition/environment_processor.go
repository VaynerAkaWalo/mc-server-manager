package definition

import (
	"context"
)

type environmentProcessor func(context.Context, map[Option]string) (map[Option]string, error)

func stringifyEnvironment(environment map[Option]string) map[string]string {
	postProcessed := make(map[string]string, len(environment))

	for key, value := range environment {
		stringKey := optsKeys[key]
		postProcessed[stringKey] = value
	}

	return postProcessed
}

func processEnvironment(ctx context.Context, environment map[Option]string) (map[Option]string, error) {
	processor := combineProcessors(ensureRequiredOpts)
	return processor(ctx, environment)
}

func combineProcessors(processors ...environmentProcessor) environmentProcessor {
	return func(ctx context.Context, environment map[Option]string) (map[Option]string, error) {
		currentEnvironment := environment
		var currentErr error

		for _, processor := range processors {
			currentEnvironment, currentErr = processor(ctx, currentEnvironment)
			if currentErr != nil {
				return nil, currentErr
			}
		}

		return currentEnvironment, nil
	}
}

func ensureRequiredOpts(ctx context.Context, environment map[Option]string) (map[Option]string, error) {
	for key, val := range requiredOpts {
		environmentValue := environment[key]
		if environmentValue == "" {
			environment[key] = val
		}
	}

	return environment, nil
}
