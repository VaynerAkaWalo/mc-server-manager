package versions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

const (
	fabricGameURL  = "https://meta.fabricmc.net/v2/versions/game"
	modrinthAPIURL = "https://api.modrinth.com/v2"
	userAgent      = "VaynerAkaWalo/mc-server-manager"

	refreshInterval = time.Hour
	retryInterval   = time.Minute
	maxBodyBytes    = 8 << 20
)

type Source struct {
	client      *http.Client
	fabricURL   string
	modrinthURL string
	mods        []string
	now         func() time.Time
	mu          sync.Mutex
	cached      definition.VersionSet
	refreshAt   time.Time
	refreshing  chan struct{}
}

func NewSource() *Source {
	return &Source{
		client:      &http.Client{Timeout: 5 * time.Second},
		fabricURL:   fabricGameURL,
		modrinthURL: modrinthAPIURL,
		mods:        definition.OptimizationMods,
		now:         time.Now,
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

type fabricGameVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type modrinthVersion struct {
	GameVersions []string `json:"game_versions"`
	VersionType  string   `json:"version_type"`
}

// Mods are optional, so a Modrinth outage keeps the mods of the previous refresh and never hides a game version.
func (s *Source) fetch(ctx context.Context, previous definition.VersionSet) (set definition.VersionSet, healthy bool, err error) {
	var fabric []fabricGameVersion
	if err := s.getJSON(ctx, s.fabricURL, &fabric); err != nil {
		return nil, false, err
	}

	modsByVersion, modsErr := s.modsByGameVersion(ctx)
	if modsErr != nil {
		slog.Error("Failed to refresh optimization mods: " + modsErr.Error())
	}

	var versions definition.TypeVersions
	for _, entry := range fabric {
		if !entry.Stable {
			continue
		}

		mods := modsByVersion[entry.Version]
		if modsErr != nil {
			kept, _ := previous[definition.ServerType].Find(entry.Version)
			mods = kept.Mods
		}

		versions.Items = append(versions.Items, definition.Version{ID: entry.Version, Kind: definition.KindRelease, Mods: mods})
	}
	if len(versions.Items) == 0 {
		return nil, false, fmt.Errorf("%s has no stable versions", s.fabricURL)
	}
	versions.LatestRelease = versions.Items[0].ID

	return definition.VersionSet{definition.ServerType: versions}, modsErr == nil, nil
}

// The server image installs only release builds of Modrinth projects, so a mod counts for a game version only when it has one.
func (s *Source) modsByGameVersion(ctx context.Context) (map[string][]string, error) {
	modsByVersion := make(map[string][]string)
	for _, mod := range s.mods {
		var releases []modrinthVersion
		endpoint := fmt.Sprintf("%s/project/%s/version?loaders=%s&include_changelog=false", s.modrinthURL, mod, url.QueryEscape(`["fabric"]`))
		if err := s.getJSON(ctx, endpoint, &releases); err != nil {
			return nil, err
		}

		supported := make(map[string]bool)
		for _, release := range releases {
			if release.VersionType != "release" {
				continue
			}
			for _, gameVersion := range release.GameVersions {
				supported[gameVersion] = true
			}
		}

		for gameVersion := range supported {
			modsByVersion[gameVersion] = append(modsByVersion[gameVersion], mod)
		}
	}

	return modsByVersion, nil
}

func (s *Source) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

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
