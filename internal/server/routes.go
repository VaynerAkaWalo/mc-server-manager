package server

import (
	"encoding/json"
	"errors"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/problem"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
	"io"
	"log/slog"
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
	router.RegisterHandler("GET /api/v1/servers", h.listServersHandler)
	router.RegisterHandler("POST /api/v1/servers", h.provisionServer)
	router.RegisterHandler("GET /api/v1/servers/stats", h.statsHandler)
	router.RegisterHandler("GET /api/v1/servers/{id}", h.getServerHandler)
	router.RegisterHandler("DELETE /api/v1/servers/{id}", h.deleteServerHandler)
	router.RegisterHandler("DELETE /servers", h.shutdownExpiredServers)
}

func (h *Handlers) listServersHandler(w http.ResponseWriter, r *http.Request) error {
	query, err := parseCollectionQuery(r.URL.Query())
	if err != nil {
		return problem.Write(w, http.StatusBadRequest, err.Error())
	}
	servers, err := h.serverService.listServers(r.Context(), query)
	if err != nil {
		return internalFailure(w, r, "Failed to get all servers", err)
	}

	return xhttp.WriteResponse(w, http.StatusOK, servers)
}

func (h *Handlers) getServerHandler(w http.ResponseWriter, r *http.Request) error {
	id, found := serverID(r)
	if !found {
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	}

	res, err := h.serverService.getServer(r.Context(), id)
	if err != nil {
		return resourceFailure(w, r, "Failed to get server", err)
	}

	return xhttp.WriteResponse(w, http.StatusOK, res)
}

func (h *Handlers) deleteServerHandler(w http.ResponseWriter, r *http.Request) error {
	id, found := serverID(r)
	if !found {
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	}

	if err := h.serverService.deleteServer(r.Context(), id); err != nil {
		return resourceFailure(w, r, "Failed to delete server", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handlers) statsHandler(w http.ResponseWriter, r *http.Request) error {
	filter, err := parseFilter(r.URL.Query())
	if err != nil {
		return problem.Write(w, http.StatusBadRequest, err.Error())
	}
	res, err := h.serverService.getStats(r.Context(), filter)
	if err != nil {
		return internalFailure(w, r, "Failed to get stats", err)
	}

	return xhttp.WriteResponse(w, http.StatusOK, res)
}

func (h *Handlers) provisionServer(w http.ResponseWriter, r *http.Request) error {
	var req server.Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err == nil && decoder.Decode(&struct{}{}) != io.EOF {
		err = errors.New("unexpected data after the JSON object")
	}
	if err != nil {
		return problem.Write(w, http.StatusBadRequest, "The request body is not valid JSON for a server")
	}

	res, err := h.serverService.provisionServer(req)
	if err != nil {
		return provisionFailure(w, r, err)
	}

	return xhttp.WriteResponse(w, http.StatusCreated, res)
}

func (h *Handlers) shutdownExpiredServers(w http.ResponseWriter, r *http.Request) error {
	results, err := h.serverService.shutdownExpiredServers()
	if err != nil {
		return internalFailure(w, r, "Failed to shut down expired servers", err)
	}

	return xhttp.WriteResponse(w, http.StatusOK, results)
}

func (h *Handlers) preFlightHandler(w http.ResponseWriter, r *http.Request) error {
	return xhttp.WriteResponse(w, http.StatusOK, nil)
}

func provisionFailure(w http.ResponseWriter, r *http.Request, err error) error {
	var invalid definition.InvalidDefinitionError
	switch {
	case errors.As(err, &invalid):
		return problem.Write(w, http.StatusUnprocessableEntity, invalid.Error())
	case errors.Is(err, ErrNameInUse):
		return problem.Write(w, http.StatusConflict, ErrNameInUse.Error())
	}

	return internalFailure(w, r, "Failed to provision server", err)
}

func internalFailure(w http.ResponseWriter, r *http.Request, message string, err error) error {
	slog.ErrorContext(r.Context(), message+": "+err.Error())

	return problem.Write(w, http.StatusInternalServerError, "Something went wrong, try again later")
}

func serverID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	return id, err == nil
}

func resourceFailure(w http.ResponseWriter, r *http.Request, message string, err error) error {
	if errors.Is(err, ErrNotFound) {
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	}

	return internalFailure(w, r, message, err)
}
