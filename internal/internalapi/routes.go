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

type Handlers struct {
	pingDatabase func(ctx context.Context) error
}

func NewHandlers(pingDatabase func(ctx context.Context) error) *Handlers {
	return &Handlers{pingDatabase: pingDatabase}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /health", h.health)
	router.RegisterHandler("GET /ready", h.ready)
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
