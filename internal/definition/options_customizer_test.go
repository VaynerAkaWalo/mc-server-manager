package definition

import (
	"context"
	"testing"
)

func TestCustomizeOptions(t *testing.T) {
	tests := []struct {
		name    string
		options map[Option]string
		wantErr bool
	}{
		{"no options", nil, false},
		{"lockout template", map[Option]string{
			VERSION:           "1.21.3",
			TYPE:              "FABRIC",
			MODRINTH_PROJECTS: "fabric-api,lithium",
			MODS:              approvedModUrls[0],
		}, false},
		{"creator form", map[Option]string{
			MOTD:        "Hello, it's my server!",
			DIFFICULTY:  "hard",
			MAX_PLAYERS: "10",
			ONLINE_MODE: "false",
			OPS:         "Steve, Alex",
		}, false},
		{"unknown mod url", map[Option]string{MODS: "https://evil.example/mod.jar"}, true},
		{"approved and unknown mod url", map[Option]string{MODS: approvedModUrls[0] + ",https://evil.example/mod.jar"}, true},
		{"enforced option", map[Option]string{MEMORY: "64Gi"}, true},
		{"jvm options", map[Option]string{JVM_XX_OPTS: "-XX:+UnlockDiagnosticVMOptions"}, true},
		{"unknown option", map[Option]string{"CUSTOM_SERVER": "https://evil.example/server.jar"}, true},
		{"unknown server type", map[Option]string{TYPE: "CUSTOM"}, true},
		{"newline in motd", map[Option]string{MOTD: "hi\nEULA=false"}, true},
		{"url in modrinth projects", map[Option]string{MODRINTH_PROJECTS: "https://evil.example/mod.jar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := customizeOptions(context.Background(), ServerDefinition{Options: tt.options})

			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err, tt.wantErr)
			}
		})
	}
}
