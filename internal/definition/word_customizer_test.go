package definition

import (
	"context"
	"testing"
	"time"
)

func TestTranslateEnforcesTierLimits(t *testing.T) {
	spec, err := TranslateDefinition(context.Background(), ServerDefinition{
		Name: "test-server", Tier: Guest, Duration: time.Hour,
		Options: map[Option]string{VERSION: "1.21.10"},
	}, testVersions)
	if err != nil {
		t.Fatal(err)
	}

	for key, want := range map[string]string{
		"MAX_PLAYERS": "3", "VIEW_DISTANCE": "6", "SIMULATION_DISTANCE": "4",
		"SYNC_CHUNK_WRITES": "false", "TYPE": "FABRIC", "MODRINTH_PROJECTS": "lithium,krypton",
	} {
		if spec.Env[key] != want {
			t.Errorf("%s = %q, want %q", key, spec.Env[key], want)
		}
	}
}
