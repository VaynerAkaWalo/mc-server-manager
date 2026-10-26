package definition

import (
	"context"
	"errors"
	"slices"
)

const (
	KindRelease = "release"

	ServerType = "FABRIC"
)

var ErrVersionsUnavailable = errors.New("the list of game versions is not available yet")

type Version struct {
	ID   string   `json:"id"`
	Kind string   `json:"kind"`
	Mods []string `json:"-"`
}

type TypeVersions struct {
	LatestRelease string    `json:"latestRelease"`
	Items         []Version `json:"items"`
}

type VersionSet map[string]TypeVersions

type VersionSource interface {
	Versions(ctx context.Context) VersionSet
}

func (v VersionSet) Versions(context.Context) VersionSet {
	return v
}

func (t TypeVersions) Find(id string) (Version, bool) {
	index := slices.IndexFunc(t.Items, func(version Version) bool { return version.ID == id })
	if index < 0 {
		return Version{}, false
	}

	return t.Items[index], true
}

func validateVersion(ctx context.Context, source VersionSource, options map[Option]string) error {
	version, requested := options[VERSION]
	if !requested {
		return nil
	}

	available := source.Versions(ctx)[ServerType]
	if len(available.Items) == 0 {
		return ErrVersionsUnavailable
	}

	if _, found := available.Find(version); !found {
		return invalidf("version %s is not available", version)
	}

	return nil
}

func customizeVersionRequired(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.Options[VERSION] == "" {
		return definition, invalidf("a game version is required")
	}

	return definition, nil
}
