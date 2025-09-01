package healthcheck

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
)

type Handlers struct {
}

func NewHealthcheckHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /health", h.healthcheckHandler)
}

func (h *Handlers) healthcheckHandler(w http.ResponseWriter, r *http.Request) error {
	return xhttp.WriteResponse(w, http.StatusOK, Response{Status: "OK"})
}
