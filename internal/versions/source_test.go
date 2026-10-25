package versions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const manifest = `{"latest":{"release":"1.21.10","snapshot":"25w41a"},"versions":[
	{"id":"25w41a","type":"snapshot"},
	{"id":"1.21.10","type":"release"},
	{"id":"1.21.9","type":"release"},
	{"id":"b1.7.3","type":"old_beta"}]}`

const fabric = `[{"version":"25w41a"},{"version":"1.21.9"}]`

func upstream(t *testing.T, failing *atomic.Bool, calls *atomic.Int32) *Source {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if failing.Load() {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if r.URL.Path == "/fabric" {
			w.Write([]byte(fabric))
			return
		}
		w.Write([]byte(manifest))
	}))
	t.Cleanup(server.Close)

	return &Source{client: server.Client(), mojangURL: server.URL + "/mojang", fabricURL: server.URL + "/fabric", now: time.Now}
}

func TestVersionsPerServerType(t *testing.T) {
	source := upstream(t, &atomic.Bool{}, &atomic.Int32{})

	set := source.Versions(context.Background())

	vanilla := set["VANILLA"]
	if vanilla.LatestRelease != "1.21.10" || len(vanilla.Items) != 3 || vanilla.Items[0].Kind != "snapshot" {
		t.Fatalf("vanilla = %+v", vanilla)
	}

	fabric := set["FABRIC"]
	if fabric.LatestRelease != "1.21.9" || len(fabric.Items) != 2 {
		t.Fatalf("fabric = %+v", fabric)
	}
}

func TestVersionsAreCachedAndSurviveOutage(t *testing.T) {
	failing, calls := &atomic.Bool{}, &atomic.Int32{}
	source := upstream(t, failing, calls)
	now := time.Now()
	source.now = func() time.Time { return now }

	source.Versions(context.Background())
	source.Versions(context.Background())
	if calls.Load() != 2 {
		t.Fatalf("upstream called %d times, want 2 for the first refresh only", calls.Load())
	}

	failing.Store(true)
	now = now.Add(2 * time.Hour)
	if got := source.Versions(context.Background())["VANILLA"].LatestRelease; got != "1.21.10" {
		t.Fatalf("stale list not served, latest = %q", got)
	}

	before := calls.Load()
	source.Versions(context.Background())
	if calls.Load() != before {
		t.Fatal("failed refresh was retried immediately")
	}
}

func TestVersionsBeforeFirstSuccessAreEmpty(t *testing.T) {
	failing := &atomic.Bool{}
	failing.Store(true)

	if set := upstream(t, failing, &atomic.Int32{}).Versions(context.Background()); len(set) != 0 {
		t.Fatalf("set = %+v", set)
	}
}
