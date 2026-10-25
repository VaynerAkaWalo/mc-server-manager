package catalog

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"net/http"
)

type Handlers struct{}

func NewCatalogHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /api/v1/catalog", h.catalogHandler)
}

func (h *Handlers) catalogHandler(w http.ResponseWriter, r *http.Request) error {
	return xhttp.WriteResponse(w, http.StatusOK, definition.BuildCatalog())
}
