package definition

import (
	"context"
	"testing"
)

func TestCustomizePresetOverridesUserOptions(t *testing.T) {
	definition := ServerDefinition{
		Preset:  "lockout",
		Options: map[Option]string{VERSION: "1.20.1", MOTD: "lockout night"},
	}

	customized, err := customizePreset(context.Background(), definition)
	if err != nil {
		t.Fatal(err)
	}

	if customized.Options[VERSION] != "1.21.3" || customized.Options[MOTD] != "lockout night" {
		t.Errorf("got options %v", customized.Options)
	}
}

func TestCustomizePresetRejectsUnknownPreset(t *testing.T) {
	definition := ServerDefinition{Preset: "my-mods", Options: map[Option]string{}}

	if _, err := customizePreset(context.Background(), definition); err == nil {
		t.Error("unknown preset was accepted")
	}
}
