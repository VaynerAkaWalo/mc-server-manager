package definition

import (
	"context"
	"testing"
)

func translatedWithSave(t *testing.T, saveID string) *ServerDefinition {
	t.Helper()
	return &ServerDefinition{Name: "my-server", Tier: Free, Options: map[Option]string{VERSION: "1.21.3"}, Duration: MaxServerLifetime, SaveId: saveID}
}

func TestTranslateDefinitionPassesTheSaveToTheCluster(t *testing.T) {
	const saveID = "4f9d0a3e-7c1b-4d6f-9a52-0e1c8b7d2a10"

	spec, err := TranslateDefinition(context.Background(), *translatedWithSave(t, saveID), testVersions)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Save == nil || spec.Save.ID != saveID || spec.Volume != "" {
		t.Fatalf("spec = %+v", spec)
	}

	spec, err = TranslateDefinition(context.Background(), *translatedWithSave(t, ""), testVersions)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Save != nil {
		t.Fatalf("spec = %+v", spec)
	}
}
