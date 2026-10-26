package definition

import (
	"context"
	"slices"
	"strings"
)

var OptimizationMods = []string{"lithium", "ferrite-core", "c2me-fabric", "scalablelux", "krypton"}

func customizePerformance(versions VersionSource) customizer {
	return func(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
		definition.Options[TYPE] = ServerType

		version, found := versions.Versions(ctx)[ServerType].Find(definition.Options[VERSION])
		if !found || len(version.Mods) == 0 {
			return definition, nil
		}

		var mods []string
		if existing := definition.Options[MODRINTH_PROJECTS]; existing != "" {
			mods = strings.Split(existing, ",")
		}

		for _, mod := range version.Mods {
			if !slices.Contains(mods, mod) {
				mods = append(mods, mod)
			}
		}
		definition.Options[MODRINTH_PROJECTS] = strings.Join(mods, ",")

		return definition, nil
	}
}
