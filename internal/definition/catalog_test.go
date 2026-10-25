package definition

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCatalogMatchesDefinitions(t *testing.T) {
	catalog := BuildCatalog()

	if len(catalog.Tiers) != len(tierMapping) {
		t.Fatalf("tiers = %+v", catalog.Tiers)
	}
	for _, tier := range catalog.Tiers {
		if _, err := translateTier(nil, Tier(tier.ID)); err != nil {
			t.Errorf("catalog tier %s cannot be provisioned: %v", tier.ID, err)
		}
	}

	if len(catalog.Presets) != 1 || catalog.Presets[0].ID != "lockout" || catalog.Presets[0].Options["VERSION"] != "1.21.3" {
		t.Fatalf("presets = %+v", catalog.Presets)
	}
	for _, option := range catalog.Options {
		for _, value := range option.Values {
			if !userSettableOpts[Option(option.ID)].valid(value) {
				t.Errorf("catalog value %s is not valid for %s", value, option.ID)
			}
		}
	}
	if catalog.DurationMs.Min != 120000 || catalog.DurationMs.Max != 1209600000 {
		t.Fatalf("duration = %+v", catalog.DurationMs)
	}
}

func TestCatalogDoesNotExposeEnforcedOptions(t *testing.T) {
	body, _ := json.Marshal(BuildCatalog())

	for _, hidden := range []string{"MODS", "MODRINTH_PROJECTS", "github.com", "MEMORY", "JVM_XX_OPTS", "EULA"} {
		if strings.Contains(string(body), hidden) {
			t.Errorf("%s is exposed: %s", hidden, body)
		}
	}
}
