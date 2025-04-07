package healthcheck

import (
	"encoding/json"
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
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(Response{Status: "OK"})
}
