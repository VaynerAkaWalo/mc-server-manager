package definition

import "slices"

type Catalog struct {
	Tiers      []CatalogTier   `json:"tiers"`
	Presets    []CatalogPreset `json:"presets"`
	Options    []CatalogOption `json:"options"`
	DurationMs DurationLimits  `json:"durationMs"`
	Versions   VersionSet      `json:"versions"`
}

type CatalogTier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type CatalogPreset struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Options     map[string]string `json:"options"`
}

type CatalogOption struct {
	ID     string   `json:"id"`
	Values []string `json:"values,omitempty"`
}

type DurationLimits struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

func BuildCatalog(versions VersionSet) Catalog {
	if versions == nil {
		versions = VersionSet{}
	}

	return Catalog{
		Versions: versions,
		Tiers:    catalogTiers(),
		Presets:  catalogPresets(),
		Options:  catalogOptions(),
		DurationMs: DurationLimits{
			Min: MinServerDuration.Milliseconds(),
			Max: MaxServerDuration.Milliseconds(),
		},
	}
}

func catalogTiers() []CatalogTier {
	tiers := make([]CatalogTier, 0, len(tierOrder))
	for _, id := range tierOrder {
		instance := tierMapping[id]
		tiers = append(tiers, CatalogTier{ID: string(id), Name: instance.name, Description: instance.description, Default: id == defaultTier})
	}

	return tiers
}

func catalogPresets() []CatalogPreset {
	result := make([]CatalogPreset, 0, len(presets))
	for id, selected := range presets {
		result = append(result, CatalogPreset{ID: id, Name: selected.name, Description: selected.description, Options: UserOptions(selected.options)})
	}
	slices.SortFunc(result, func(a, b CatalogPreset) int { return compareStrings(a.ID, b.ID) })

	return result
}

func catalogOptions() []CatalogOption {
	result := make([]CatalogOption, 0, len(userSettableOpts))
	for id, spec := range userSettableOpts {
		result = append(result, CatalogOption{ID: string(id), Values: spec.values})
	}
	slices.SortFunc(result, func(a, b CatalogOption) int { return compareStrings(a.ID, b.ID) })

	return result
}

func compareStrings(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
