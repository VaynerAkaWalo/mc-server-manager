package server

import (
	"encoding/json"
	"fmt"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
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

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /servers", h.listServersHandler)
	router.RegisterHandler("POST /servers", h.provisionServer)
	router.RegisterHandler("DELETE /servers", h.shutdownExpiredServers)
}

func (h *Handlers) listServersHandler(w http.ResponseWriter, r *http.Request) error {
	servers, err := h.serverService.getActiveServers()
	if err != nil {
		return xhttp.NewError("Failed to get all servers "+err.Error(), http.StatusInternalServerError)
	}

	return xhttp.WriteResponse(w, http.StatusOK, servers)
}

func (h *Handlers) provisionServer(w http.ResponseWriter, r *http.Request) error {
	var req server.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" || req.Duration == 0 {
		return xhttp.NewError(fmt.Sprintf("Invalid json structure %v", err), http.StatusBadRequest)
	}

	res, err := h.serverService.provisionServer(req)
	if err != nil {
		return xhttp.NewError(err.Error(), http.StatusInternalServerError)
	}

	return xhttp.WriteResponse(w, http.StatusCreated, res)
}

func (h *Handlers) shutdownExpiredServers(w http.ResponseWriter, r *http.Request) error {
	results, err := h.serverService.shutdownExpiredServers()
	if err != nil {
		return xhttp.NewError(err.Error(), http.StatusInternalServerError)
	}

	return xhttp.WriteResponse(w, http.StatusOK, results)
}

func (h *Handlers) preFlightHandler(w http.ResponseWriter, r *http.Request) error {
	return xhttp.WriteResponse(w, http.StatusOK, nil)
}
