package internalapi

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/problem"
	"log/slog"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type expirer interface {
	ShutdownExpired() ([]string, error)
}

type Handlers struct {
	pingDatabase func(ctx context.Context) error
	servers      expirer
}

func NewHandlers(pingDatabase func(ctx context.Context) error, servers expirer) *Handlers {
	return &Handlers{
		pingDatabase: pingDatabase,
		servers:      servers,
	}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /health", h.health)
	router.RegisterHandler("GET /ready", h.ready)
	router.RegisterHandler("POST /internal/v1/jobs/expire-servers", h.expireServers)
}

func (h *Handlers) health(w http.ResponseWriter, r *http.Request) error {
	return xhttp.WriteResponse(w, http.StatusOK, statusResponse{Status: "OK"})
}

func (h *Handlers) ready(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()

	if err := h.pingDatabase(ctx); err != nil {
		slog.ErrorContext(r.Context(), "Database is not reachable: "+err.Error())
		return problem.Write(w, http.StatusServiceUnavailable, "The service is not ready")
	}

	return xhttp.WriteResponse(w, http.StatusOK, statusResponse{Status: "OK"})
}

// Running it again after a run finished, or while servers are still shutting down, is safe and expires nothing twice.
func (h *Handlers) expireServers(w http.ResponseWriter, r *http.Request) error {
	names, err := h.servers.ShutdownExpired()
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to expire servers: "+err.Error())
		return problem.Write(w, http.StatusInternalServerError, "Something went wrong, try again later")
	}

	return xhttp.WriteResponse(w, http.StatusOK, expireServersResponse{Expired: names})
}
