package definition

import (
	"context"
	"errors"
	"slices"
)

const (
	KindRelease  = "release"
	KindSnapshot = "snapshot"

	defaultServerType = "VANILLA"
)

var ErrVersionsUnavailable = errors.New("the list of game versions is not available yet")

type Version struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type TypeVersions struct {
	LatestRelease string    `json:"latestRelease"`
	Items         []Version `json:"items"`
}

// VersionSet maps a server type to the game versions it can run.
type VersionSet map[string]TypeVersions

type VersionSource interface {
	Versions(ctx context.Context) VersionSet
}

func (v VersionSet) Versions(context.Context) VersionSet {
	return v
}

func (t TypeVersions) contains(id string) bool {
	return slices.ContainsFunc(t.Items, func(version Version) bool { return version.ID == id })
}

func validateVersion(ctx context.Context, source VersionSource, options map[Option]string) error {
	version, requested := options[VERSION]
	if !requested {
		return nil
	}

	serverType := options[TYPE]
	if serverType == "" {
		serverType = defaultServerType
	}

	available := source.Versions(ctx)[serverType]
	if len(available.Items) == 0 {
		return ErrVersionsUnavailable
	}

	if !available.contains(version) {
		return invalidf("version %s is not available for server type %s", version, serverType)
	}

	return nil
}
