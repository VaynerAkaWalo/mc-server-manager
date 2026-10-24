package frontend

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestServeApp(t *testing.T) {
	files := fstest.MapFS{
		"index.html":    {Data: []byte("app")},
		"assets/app.js": {Data: []byte("script")},
	}

	mux := http.NewServeMux()
	NewFrontendHandlers(files).RegisterRoutes(&xhttp.Router{ServeMux: mux})

	indexPaths := map[string]bool{"/": true, "/server-creator": true, "/assets": true}

	tests := map[string]string{
		"/":               "app",
		"/server-creator": "app",
		"/assets/app.js":  "script",
		"/assets":         "app",
	}

	for url, want := range tests {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))

		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Errorf("GET %s = %d %q, want 200 %q", url, rec.Code, rec.Body.String(), want)
		}

		if got := rec.Header().Get("Cache-Control"); indexPaths[url] != (got == "no-cache") {
			t.Errorf("GET %s Cache-Control = %q", url, got)
		}
	}
}
