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
	name         string
	description  string
	memory       string
	instanceType string
}

// Game nodes have about 6.5Gi allocatable, so two wooden servers share a node and an iron server takes a whole one.
const (
	MemorySmall  string = "3000Mi"
	MemoryMedium string = "6000Mi"
	Shared       string = "shared"
)

var tierMapping = map[Tier]serverInstance{
	Wooden: {name: "Wooden", description: "When chunks are already generated", memory: MemorySmall, instanceType: Shared},
	Iron:   {name: "Iron", description: "Default option, should handle up to 12 players", memory: MemoryMedium, instanceType: Shared},
}

const defaultTier = Iron

var tierOrder = []Tier{Wooden, Iron}

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
