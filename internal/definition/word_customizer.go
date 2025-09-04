package definition

import (
	"context"
	"strconv"
)

const DefaultViewDistance = "10"

func customizeWorldGeneration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	val, found := definition.Options[VIEW_DISTANCE]
	if !found {
		definition.Options[VIEW_DISTANCE] = DefaultViewDistance
	}

	viewDistance, err := strconv.Atoi(val)
	if err == nil || viewDistance < 1 || viewDistance > 64 {
		definition.Options[VIEW_DISTANCE] = DefaultViewDistance
	}

	return definition, nil
}
