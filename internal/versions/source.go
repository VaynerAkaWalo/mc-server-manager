package versions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

const (
	mojangManifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
	fabricGameURL     = "https://meta.fabricmc.net/v2/versions/game"

	refreshInterval = time.Hour
	retryInterval   = time.Minute
	maxBodyBytes    = 8 << 20
)

type Source struct {
	client    *http.Client
	mojangURL string
	fabricURL string
	now       func() time.Time
	mu        sync.Mutex
	cached    definition.VersionSet
	refreshAt time.Time
}

func NewSource() *Source {
	return &Source{
		client:    &http.Client{Timeout: 5 * time.Second},
		mojangURL: mojangManifestURL,
		fabricURL: fabricGameURL,
		now:       time.Now,
	}
}

// Versions serves the last known list when the upstream is down, and an empty one before the first success.
func (s *Source) Versions(ctx context.Context) definition.VersionSet {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.now().Before(s.refreshAt) {
		return s.cached
	}

	fetched, err := s.fetch(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to refresh game versions: "+err.Error())
		s.refreshAt = s.now().Add(retryInterval)

		return s.cached
	}

	s.cached = fetched
	s.refreshAt = s.now().Add(refreshInterval)

	return s.cached
}

type mojangManifest struct {
	Latest struct {
		Release string `json:"release"`
	} `json:"latest"`
	Versions []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"versions"`
}

type fabricGameVersion struct {
	Version string `json:"version"`
}

func (s *Source) fetch(ctx context.Context) (definition.VersionSet, error) {
	var manifest mojangManifest
	if err := s.getJSON(ctx, s.mojangURL, &manifest); err != nil {
		return nil, err
	}

	var fabricVersions []fabricGameVersion
	if err := s.getJSON(ctx, s.fabricURL, &fabricVersions); err != nil {
		return nil, err
	}

	supportedByFabric := make(map[string]bool, len(fabricVersions))
	for _, version := range fabricVersions {
		supportedByFabric[version.Version] = true
	}

	vanilla := definition.TypeVersions{LatestRelease: manifest.Latest.Release}
	fabric := definition.TypeVersions{}
	for _, entry := range manifest.Versions {
		kind, listed := kindOf(entry.Type)
		if !listed {
			continue
		}

		version := definition.Version{ID: entry.ID, Kind: kind}
		vanilla.Items = append(vanilla.Items, version)
		if !supportedByFabric[entry.ID] {
			continue
		}

		fabric.Items = append(fabric.Items, version)
		if fabric.LatestRelease == "" && kind == definition.KindRelease {
			fabric.LatestRelease = entry.ID
		}
	}

	if len(vanilla.Items) == 0 {
		return nil, fmt.Errorf("manifest %s has no versions", s.mojangURL)
	}

	return definition.VersionSet{"VANILLA": vanilla, "FABRIC": fabric}, nil
}

func kindOf(mojangType string) (string, bool) {
	switch mojangType {
	case "release":
		return definition.KindRelease, true
	case "snapshot":
		return definition.KindSnapshot, true
	}

	return "", false
}

func (s *Source) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned %d", url, res.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(res.Body, maxBodyBytes)).Decode(target)
}
