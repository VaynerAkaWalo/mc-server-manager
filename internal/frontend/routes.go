package frontend

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type Handlers struct {
	files      fs.FS
	fileServer http.Handler
}

func NewFrontendHandlers(files fs.FS) *Handlers {
	return &Handlers{
		files:      files,
		fileServer: http.FileServerFS(files),
	}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.Handle("GET /", http.HandlerFunc(h.serveApp))
}

func (h *Handlers) serveApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if h.isFile(name) {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	// Hashed assets from older builds disappear on deploy, so browsers must revalidate the index that references them.
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFileFS(w, r, h.files, "index.html")
}

func (h *Handlers) isFile(name string) bool {
	info, err := fs.Stat(h.files, name)
	return err == nil && !info.IsDir()
}
