package definition

import (
	"context"
	"strconv"
)

func customizeWorldGeneration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	instance, err := translateTier(ctx, definition.Tier)
	if err != nil {
		return definition, err
	}

	definition.Options[MAX_PLAYERS] = strconv.Itoa(instance.maxPlayers)
	definition.Options[VIEW_DISTANCE] = strconv.Itoa(instance.viewDistance)
	definition.Options[SIMULATION_DISTANCE] = strconv.Itoa(instance.simulationDistance)
	definition.Options[SYNC_CHUNK_WRITES] = "false"

	return definition, nil
}
