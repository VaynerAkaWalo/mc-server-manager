package catalog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

func TestCatalogRoute(t *testing.T) {
	mux := http.NewServeMux()
	NewCatalogHandlers().RegisterRoutes(&xhttp.Router{ServeMux: mux})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/catalog", nil))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"tiers":[{"id":"wooden"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
