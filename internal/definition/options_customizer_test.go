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
		{"creator form", map[Option]string{
			MOTD:        "Hello, it's my server!",
			DIFFICULTY:  "hard",
			MAX_PLAYERS: "10",
			ONLINE_MODE: "false",
			OPS:         "Steve, Alex",
		}, false},
		{"mod url", map[Option]string{MODS: presets["lockout"].options[MODS]}, true},
		{"modrinth projects", map[Option]string{MODRINTH_PROJECTS: "fabric-api"}, true},
		{"enforced option", map[Option]string{MEMORY: "64Gi"}, true},
		{"jvm options", map[Option]string{JVM_XX_OPTS: "-XX:+UnlockDiagnosticVMOptions"}, true},
		{"unknown option", map[Option]string{"CUSTOM_SERVER": "https://evil.example/server.jar"}, true},
		{"unknown server type", map[Option]string{TYPE: "CUSTOM"}, true},
		{"newline in motd", map[Option]string{MOTD: "hi\nEULA=false"}, true},
		{"url in version", map[Option]string{VERSION: "https://evil.example/server.jar"}, true},
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

func TestCustomizeOptionsRejectsValuesOutsideEnums(t *testing.T) {
	for _, options := range []map[Option]string{{DIFFICULTY: "impossible"}, {ONLINE_MODE: "maybe"}} {
		if _, err := customizeOptions(context.Background(), ServerDefinition{Options: options}); err == nil {
			t.Errorf("%v was accepted", options)
		}
	}
}
