package definition

import (
	"context"
	"testing"
)

func TestCustomizePerformance(t *testing.T) {
	tests := []struct {
		name     string
		options  map[Option]string
		wantMods string
	}{
		{"mods with a release for the version", map[Option]string{VERSION: "1.21.10"}, "lithium,krypton"},
		{"no mods for the version", map[Option]string{VERSION: "1.21.3"}, ""},
		{"unknown version", map[Option]string{VERSION: "1.99"}, ""},
		{"preset mods", map[Option]string{VERSION: "1.21.10", MODRINTH_PROJECTS: "fabric-api,lithium"}, "fabric-api,lithium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := customizePerformance(testVersions)(context.Background(), ServerDefinition{Options: tt.options})
			if err != nil {
				t.Fatal(err)
			}

			if got.Options[TYPE] != ServerType || got.Options[MODRINTH_PROJECTS] != tt.wantMods {
				t.Fatalf("options = %+v", got.Options)
			}
		})
	}
}
