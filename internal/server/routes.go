package server

import (
	"encoding/json"
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
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.SendErrorResponse(w, http.StatusBadRequest, "Invalid json structure")
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
