package server

import (
	"encoding/json"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/web"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
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
	router.HandleFunc("OPTIONS /servers", h.preFlightHandler)
	router.HandleFunc("GET /servers", h.listServersHandler)
	router.HandleFunc("POST /servers", h.provisionServer)
	router.HandleFunc("DELETE /servers", h.shutdownExpiredServers)
}

func (h *Handlers) listServersHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := h.serverService.getActiveServers()
	if err != nil {
		web.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get all servers "+err.Error())
		return
	}

	web.SendJsonResponse(w, http.StatusOK, servers)
}

func (h *Handlers) provisionServer(w http.ResponseWriter, r *http.Request) {
	var req server.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" || req.ExpireAfter == 0 {
		web.SendErrorResponse(w, http.StatusBadRequest, "Invalid json structure "+err.Error())
		return
	}

	res, err := h.serverService.provisionServer(req)
	if err != nil {
		web.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	web.SendJsonResponse(w, http.StatusCreated, res)
}

func (h *Handlers) shutdownExpiredServers(w http.ResponseWriter, r *http.Request) {
	results, err := h.serverService.shutdownExpiredServers()
	if err != nil {
		web.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
	}

	web.SendJsonResponse(w, http.StatusOK, results)
}

func (h *Handlers) preFlightHandler(w http.ResponseWriter, r *http.Request) {
	web.SendJsonResponse(w, http.StatusOK, nil)
}
