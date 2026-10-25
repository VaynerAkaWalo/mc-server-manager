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
	client     *http.Client
	mojangURL  string
	fabricURL  string
	now        func() time.Time
	mu         sync.Mutex
	cached     definition.VersionSet
	refreshAt  time.Time
	refreshing chan struct{}
}

func NewSource() *Source {
	return &Source{
		client:    &http.Client{Timeout: 5 * time.Second},
		mojangURL: mojangManifestURL,
		fabricURL: fabricGameURL,
		now:       time.Now,
	}
}

func (s *Source) Versions(ctx context.Context) definition.VersionSet {
	s.mu.Lock()
	cached := s.cached
	if s.now().Before(s.refreshAt) {
		s.mu.Unlock()
		return cached
	}

	done := s.refreshing
	if done == nil {
		done = make(chan struct{})
		s.refreshing = done
		go s.refresh(cached, done)
	}
	s.mu.Unlock()

	if len(cached) > 0 {
		return cached
	}

	select {
	case <-done:
	case <-ctx.Done():
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cached
}

// The refresh outlives the request that triggered it, so a disconnecting client cannot fail it for everyone.
func (s *Source) refresh(previous definition.VersionSet, done chan struct{}) {
	ctx := context.Background()
	fetched, healthy, err := s.fetch(ctx, previous)
	if err != nil {
		slog.Error("Failed to refresh game versions: " + err.Error())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	interval := retryInterval
	if err == nil {
		s.cached = fetched
		if healthy {
			interval = refreshInterval
		}
	}
	s.refreshAt = s.now().Add(interval)
	s.refreshing = nil
	close(done)
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

// A Fabric outage must not hide new vanilla versions, so each type keeps its previous list when its own source fails.
func (s *Source) fetch(ctx context.Context, previous definition.VersionSet) (set definition.VersionSet, healthy bool, err error) {
	var manifest mojangManifest
	if err := s.getJSON(ctx, s.mojangURL, &manifest); err != nil {
		return nil, false, err
	}

	vanilla := definition.TypeVersions{LatestRelease: manifest.Latest.Release}
	for _, entry := range manifest.Versions {
		if kind, listed := kindOf(entry.Type); listed {
			vanilla.Items = append(vanilla.Items, definition.Version{ID: entry.ID, Kind: kind})
		}
	}
	if len(vanilla.Items) == 0 {
		return nil, false, fmt.Errorf("manifest %s has no versions", s.mojangURL)
	}

	set = definition.VersionSet{"VANILLA": vanilla}

	fabric, err := s.fabricVersions(ctx, vanilla)
	if err != nil {
		slog.Error("Failed to refresh Fabric game versions: " + err.Error())
		if kept, found := previous["FABRIC"]; found {
			set["FABRIC"] = kept
		}

		return set, false, nil
	}

	set["FABRIC"] = fabric

	return set, true, nil
}

func (s *Source) fabricVersions(ctx context.Context, vanilla definition.TypeVersions) (definition.TypeVersions, error) {
	var supported []fabricGameVersion
	if err := s.getJSON(ctx, s.fabricURL, &supported); err != nil {
		return definition.TypeVersions{}, err
	}

	supportedByFabric := make(map[string]bool, len(supported))
	for _, version := range supported {
		supportedByFabric[version.Version] = true
	}

	var fabric definition.TypeVersions
	for _, version := range vanilla.Items {
		if !supportedByFabric[version.ID] {
			continue
		}

		fabric.Items = append(fabric.Items, version)
		if fabric.LatestRelease == "" && version.Kind == definition.KindRelease {
			fabric.LatestRelease = version.ID
		}
	}
	if len(fabric.Items) == 0 {
		return definition.TypeVersions{}, fmt.Errorf("%s has no versions known to Mojang", s.fabricURL)
	}

	return fabric, nil
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
