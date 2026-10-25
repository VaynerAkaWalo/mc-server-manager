package catalog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
)

func TestCatalogRoute(t *testing.T) {
	mux := http.NewServeMux()
	NewCatalogHandlers(definition.VersionSet{"VANILLA": {LatestRelease: "1.21.10"}}).RegisterRoutes(&xhttp.Router{ServeMux: mux})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/catalog", nil))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"tiers":[{"id":"wooden"`) || !strings.Contains(rec.Body.String(), `"latestRelease":"1.21.10"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
