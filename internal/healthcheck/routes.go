package healthcheck

import (
	"github.com/VaynerAkaWalo/mc-server-manager/internal/web"
	"net/http"
)

type Handlers struct {
}

func NewHealthcheckHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) RegisterHealthcheckRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /health", h.healthcheckHandler)
}

func (h *Handlers) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	web.SendJsonResponse(w, http.StatusOK, Response{Status: "OK"})
}
