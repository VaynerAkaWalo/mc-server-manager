package definition

import (
	"context"
	"errors"
	"testing"
	"time"
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
		{"known version", map[Option]string{VERSION: "1.21.3"}, false},
		{"snapshot", map[Option]string{VERSION: "25w41a"}, false},
		{"unknown version", map[Option]string{VERSION: "1.99"}, true},
		{"version of another type", map[Option]string{TYPE: "FABRIC", VERSION: "1.21.10"}, true},
		{"url in version", map[Option]string{VERSION: "https://evil.example/server.jar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := customizeOptions(testVersions)(context.Background(), ServerDefinition{Options: tt.options})

			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomizeOptionsRejectsValuesOutsideEnums(t *testing.T) {
	for _, options := range []map[Option]string{{DIFFICULTY: "impossible"}, {ONLINE_MODE: "maybe"}} {
		if _, err := customizeOptions(testVersions)(context.Background(), ServerDefinition{Options: options}); err == nil {
			t.Errorf("%v was accepted", options)
		}
	}
}

var testVersions = VersionSet{
	"VANILLA": {LatestRelease: "1.21.10", Items: []Version{{"1.21.10", KindRelease}, {"25w41a", KindSnapshot}, {"1.21.3", KindRelease}}},
	"FABRIC":  {LatestRelease: "1.21.3", Items: []Version{{"1.21.3", KindRelease}}},
}

func TestCustomizeOptionsWithoutVersionList(t *testing.T) {
	_, err := customizeOptions(VersionSet{})(context.Background(), ServerDefinition{Options: map[Option]string{VERSION: "1.21.3"}})

	if !errors.Is(err, ErrVersionsUnavailable) {
		t.Fatalf("got error %v", err)
	}
}

func TestVersionIsRequiredUnlessPresetSetsIt(t *testing.T) {
	_, err := customizeVersionRequired(context.Background(), ServerDefinition{Options: map[Option]string{}})
	if err == nil {
		t.Fatal("definition without version was accepted")
	}

	withPreset, err := customizePreset(context.Background(), ServerDefinition{Preset: "lockout", Options: map[Option]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := customizeVersionRequired(context.Background(), withPreset); err != nil {
		t.Fatalf("preset version was rejected: %v", err)
	}
}

func TestTranslateRequiresVersionAfterPreset(t *testing.T) {
	base := ServerDefinition{Name: "my-server", Tier: Wooden, Duration: time.Hour, Options: map[Option]string{}}

	if _, err := TranslateDefinition(context.Background(), base, testVersions); err == nil {
		t.Fatal("definition without version was translated")
	}

	base.Preset = "lockout"
	if _, err := TranslateDefinition(context.Background(), base, testVersions); err != nil {
		t.Fatalf("preset without explicit version was rejected: %v", err)
	}
}
