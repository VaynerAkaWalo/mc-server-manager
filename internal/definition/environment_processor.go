package definition

import (
	"context"
)

const adminOp = "554156b6-2a93-3fe2-a63e-45f4fa95ec35,510bd8128b1b4a5cacd80143a76cab51"

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
	processor := combineProcessors(ensureRequiredOpts, ensureAdminsHaveOperator)
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

func ensureAdminsHaveOperator(ctx context.Context, environment map[Option]string) (map[Option]string, error) {
	currentOps := environment[OPS]
	if currentOps == "" {
		environment[OPS] = adminOp
	} else {
		environment[OPS] = currentOps + "," + adminOp
	}

	return environment, nil
}
