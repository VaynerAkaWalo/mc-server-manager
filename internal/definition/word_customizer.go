package definition

import (
	"context"
	"strconv"
)

const DefaultViewDistance = "10"

var tierWorldOptions = map[Tier]map[Option]string{
	Wooden: {VIEW_DISTANCE: DefaultViewDistance, SIMULATION_DISTANCE: "6", MAX_PLAYERS: "5", SYNC_CHUNK_WRITES: "false"},
	Iron:   {VIEW_DISTANCE: DefaultViewDistance, SIMULATION_DISTANCE: "8", MAX_PLAYERS: "12", SYNC_CHUNK_WRITES: "false"},
}

func customizeWorldGeneration(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	for key, value := range tierWorldOptions[definition.Tier] {
		if definition.Options[key] == "" {
			definition.Options[key] = value
		}
	}

	viewDistance, err := strconv.Atoi(definition.Options[VIEW_DISTANCE])
	if err != nil || viewDistance < 1 || viewDistance > 64 {
		definition.Options[VIEW_DISTANCE] = DefaultViewDistance
	}

	return definition, nil
}
