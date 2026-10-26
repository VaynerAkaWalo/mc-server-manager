package definition

import (
	"context"
	"testing"
	"time"
)

func TestTranslateTierWorldDefaults(t *testing.T) {
	for _, tt := range []struct {
		tier       Tier
		players    string
		simulation string
	}{
		{Wooden, "5", "6"},
		{Iron, "12", "8"},
	} {
		t.Run(string(tt.tier), func(t *testing.T) {
			spec, err := TranslateDefinition(context.Background(), ServerDefinition{
				Name: "test-server", Tier: tt.tier, Duration: time.Hour,
				Options: map[Option]string{VERSION: "1.21.10"},
			}, testVersions)
			if err != nil {
				t.Fatal(err)
			}

			for key, want := range map[string]string{
				"VIEW_DISTANCE": "10", "SIMULATION_DISTANCE": tt.simulation,
				"MAX_PLAYERS": tt.players, "SYNC_CHUNK_WRITES": "false",
				"TYPE": "FABRIC", "MODRINTH_PROJECTS": "lithium,krypton",
			} {
				if spec.Env[key] != want {
					t.Errorf("%s = %q, want %q", key, spec.Env[key], want)
				}
			}
		})
	}
}

func TestWorldOptionsPreserveOverrides(t *testing.T) {
	for _, view := range []string{"4", "64", "0", "65", "invalid"} {
		t.Run(view, func(t *testing.T) {
			got, err := customizeWorldGeneration(context.Background(), ServerDefinition{
				Tier: Wooden, Options: map[Option]string{VIEW_DISTANCE: view, MAX_PLAYERS: "3"},
			})
			if err != nil {
				t.Fatal(err)
			}

			wantView := DefaultViewDistance
			if view == "4" || view == "64" {
				wantView = view
			}
			if got.Options[VIEW_DISTANCE] != wantView || got.Options[MAX_PLAYERS] != "3" {
				t.Fatalf("options = %+v", got.Options)
			}
		})
	}
}
