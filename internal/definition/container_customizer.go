package definition

import (
	"context"
)

var containerEnforcedOpts = map[Option]string{
	MEMORY:          "",
	USE_AIKAR_FLAGS: "true",
	JVM_XX_OPTS:     "-XX:MaxRAMPercentage=75",
}

type serverInstance struct {
	name               string
	description        string
	memory             string
	instanceType       string
	maxPlayers         int
	viewDistance       int
	simulationDistance int
	persistent         bool
}

// Game nodes have about 6.5Gi allocatable, so two small servers share a node.
const (
	MemorySmall string = "3000Mi"
	Shared      string = "shared"
)

var tierMapping = map[Tier]serverInstance{
	Guest: {
		name:               "Guest",
		description:        "For playing without an account",
		memory:             MemorySmall,
		instanceType:       Shared,
		maxPlayers:         3,
		viewDistance:       6,
		simulationDistance: 4,
		persistent:         false,
	},
}

const defaultTier = Guest

var tierOrder = []Tier{Guest}

func customizeContainer(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	for key, val := range containerEnforcedOpts {
		definition.Options[key] = val
	}

	return definition, nil
}

func translateTier(ctx context.Context, tier Tier) (serverInstance, error) {
	instance, found := tierMapping[tier]
	if !found {
		return serverInstance{}, invalidf("tier %s is not available", tier)
	}

	return instance, nil
}
