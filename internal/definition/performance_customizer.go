package definition

import (
	"context"
	"strings"
)

// Every server runs Fabric so these server-side mods can be added, while players still join with a vanilla client.
var OptimizationMods = []string{"lithium", "ferrite-core", "c2me-fabric", "scalablelux", "krypton"}

func customizePerformance(versions VersionSource) customizer {
	return func(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
		definition.Options[TYPE] = ServerType

		if definition.Options[MODRINTH_PROJECTS] != "" {
			return definition, nil
		}

		version, found := versions.Versions(ctx)[ServerType].Find(definition.Options[VERSION])
		if found && len(version.Mods) > 0 {
			definition.Options[MODRINTH_PROJECTS] = strings.Join(version.Mods, ",")
		}

		return definition, nil
	}
}
