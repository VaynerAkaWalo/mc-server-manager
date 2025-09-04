package definition

import "context"

var containerEnforcedOpts = map[Option]string{
	MEMORY:          "",
	USE_AIKAR_FLAGS: "true",
	JVM_XX_OPTS:     "-XX:MaxRAMPercentage=75",
}

type serverInstance struct {
	memory       string
	instanceType string
}

const (
	MemorySmall  string = "3300Mi"
	MemoryMedium string = "7300Mi"
	MemoryLarge  string = "15300Mi"
	Shared       string = "shared"
)

var tierMapping = map[Tier]serverInstance{
	Wooden:  {memory: MemorySmall, instanceType: Shared},
	Iron:    {memory: MemoryMedium, instanceType: Shared},
	Diamond: {memory: MemoryLarge, instanceType: Shared},
}

func customizeContainer(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	for key, val := range containerEnforcedOpts {
		definition.Options[key] = val
	}

	return definition, nil
}

func translateTier(ctx context.Context, tier Tier) serverInstance {
	return tierMapping[tier]
}
