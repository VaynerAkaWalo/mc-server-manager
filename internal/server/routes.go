package server

import (
	"github.com/VaynerAkaWalo/mc-server-manager/internal/web"
	"net/http"
)

type Handlers struct {
	serverService Service
}

func NewServerHandlers(serverService Service) *Handlers {
	return &Handlers{
		serverService: serverService,
	}
}

func (h *Handlers) RegisterServerRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /servers", h.listServersHandler)
}

func (h *Handlers) listServersHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := h.serverService.getActiveServers()
	if err != nil {
		web.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get all servers "+err.Error())
	}

	web.SendJsonResponse(w, http.StatusOK, servers)
}
