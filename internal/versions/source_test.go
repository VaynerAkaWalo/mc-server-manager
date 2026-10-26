package versions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

const fabric = `[
	{"version":"25w41a","stable":false},
	{"version":"1.21.10","stable":true},
	{"version":"1.21.9","stable":true}]`

var modReleases = map[string]string{
	"lithium": `[{"version_type":"release","game_versions":["1.21.10","1.21.9"]}]`,
	"c2me":    `[{"version_type":"alpha","game_versions":["1.21.10"]},{"version_type":"release","game_versions":["1.21.9"]}]`,
}

func upstream(t *testing.T, fabricDown, modrinthDown *atomic.Bool, calls *atomic.Int32) *Source {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/fabric" {
			if fabricDown.Load() {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.Write([]byte(fabric))
			return
		}

		mod := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/modrinth/project/"), "/version")
		if modrinthDown.Load() || r.URL.Query().Get("loaders") != `["fabric"]` {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Write([]byte(modReleases[mod]))
	}))
	t.Cleanup(server.Close)

	return &Source{
		client:      server.Client(),
		fabricURL:   server.URL + "/fabric",
		modrinthURL: server.URL + "/modrinth",
		mods:        []string{"lithium", "c2me"},
		now:         time.Now,
	}
}

func TestVersionsAreStableFabricReleasesWithTheirMods(t *testing.T) {
	source := upstream(t, &atomic.Bool{}, &atomic.Bool{}, &atomic.Int32{})

	versions := source.Versions(context.Background())[definition.ServerType]

	want := definition.TypeVersions{LatestRelease: "1.21.10", Items: []definition.Version{
		{ID: "1.21.10", Kind: definition.KindRelease, Mods: []string{"lithium"}},
		{ID: "1.21.9", Kind: definition.KindRelease, Mods: []string{"lithium", "c2me"}},
	}}
	if !reflect.DeepEqual(versions, want) {
		t.Fatalf("versions = %+v", versions)
	}
}

func settle(source *Source) {
	source.mu.Lock()
	done := source.refreshing
	source.mu.Unlock()

	if done != nil {
		<-done
	}
}

func TestVersionsAreCachedAndSurviveOutage(t *testing.T) {
	failing, calls := &atomic.Bool{}, &atomic.Int32{}
	source := upstream(t, failing, &atomic.Bool{}, calls)
	now := time.Now()
	source.now = func() time.Time { return now }

	source.Versions(context.Background())
	source.Versions(context.Background())
	if calls.Load() != 3 {
		t.Fatalf("upstream called %d times, want 3 for the first refresh only", calls.Load())
	}

	failing.Store(true)
	now = now.Add(2 * time.Hour)
	if got := latest(source); got != "1.21.10" {
		t.Fatalf("stale list not served, latest = %q", got)
	}
	settle(source)

	before := calls.Load()
	if got := latest(source); got != "1.21.10" {
		t.Fatalf("stale list lost after failed refresh, latest = %q", got)
	}
	settle(source)
	if calls.Load() != before {
		t.Fatal("failed refresh was retried immediately")
	}
}

func TestCancelledCallerDoesNotBlockOthers(t *testing.T) {
	source := upstream(t, &atomic.Bool{}, &atomic.Bool{}, &atomic.Int32{})
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	source.Versions(cancelled)
	settle(source)

	if got := latest(source); got != "1.21.10" {
		t.Fatalf("latest = %q", got)
	}
}

func TestModrinthOutageKeepsPreviousModsAndRetriesSoon(t *testing.T) {
	modrinthDown := &atomic.Bool{}
	source := upstream(t, &atomic.Bool{}, modrinthDown, &atomic.Int32{})
	now := time.Now()
	source.now = func() time.Time { return now }

	modrinthDown.Store(true)
	cold := source.Versions(context.Background())[definition.ServerType]
	if len(cold.Items) != 2 || cold.Items[0].Mods != nil {
		t.Fatalf("cold start = %+v", cold)
	}
	if !source.refreshAt.Equal(now.Add(retryInterval)) {
		t.Fatal("refresh without mods is not retried soon")
	}

	modrinthDown.Store(false)
	source.refreshAt = time.Time{}
	source.Versions(context.Background())
	settle(source)
	modrinthDown.Store(true)
	source.refreshAt = time.Time{}
	source.Versions(context.Background())
	settle(source)

	if got := source.Versions(context.Background())[definition.ServerType].Items[1].Mods; !reflect.DeepEqual(got, []string{"lithium", "c2me"}) {
		t.Fatalf("previous mods not kept, mods = %v", got)
	}
}

func TestVersionsBeforeFirstSuccessAreEmpty(t *testing.T) {
	failing := &atomic.Bool{}
	failing.Store(true)

	if set := upstream(t, failing, &atomic.Bool{}, &atomic.Int32{}).Versions(context.Background()); len(set) != 0 {
		t.Fatalf("set = %+v", set)
	}
}

func latest(source *Source) string {
	return source.Versions(context.Background())[definition.ServerType].LatestRelease
}
