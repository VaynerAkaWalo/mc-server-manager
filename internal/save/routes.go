package save

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/auth"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/problem"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Request struct {
	Name string `json:"name"`
}

type Collection struct {
	Items      []Response `json:"items"`
	NextCursor string     `json:"nextCursor"`
}

type Authenticator interface {
	Authenticated(next auth.SessionHandler) func(w http.ResponseWriter, r *http.Request) error
}

type Handlers struct {
	service *Service
	auth    Authenticator
}

func NewHandlers(service *Service, auth Authenticator) *Handlers {
	return &Handlers{service: service, auth: auth}
}

func (h *Handlers) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /api/v1/saves", h.auth.Authenticated(h.list))
	router.RegisterHandler("POST /api/v1/saves", h.auth.Authenticated(h.create))
	router.RegisterHandler("PATCH /api/v1/saves/{id}", h.auth.Authenticated(h.rename))
	router.RegisterHandler("DELETE /api/v1/saves/{id}", h.auth.Authenticated(h.delete))
}

func newResponse(record store.SaveRecord) Response {
	return Response{ID: record.ID, Name: record.Name, CreatedAt: record.CreatedAt}
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request, session store.Session) error {
	records, err := h.service.List(r.Context(), session.Subject)
	if err != nil {
		return failure(w, r, "Failed to list saves", err)
	}

	collection := Collection{Items: make([]Response, 0, len(records))}
	for _, record := range records {
		collection.Items = append(collection.Items, newResponse(record))
	}

	return xhttp.WriteResponse(w, http.StatusOK, collection)
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request, session store.Session) error {
	request, ok := decode(w, r)
	if !ok {
		return nil
	}

	record, err := h.service.Create(r.Context(), session.Subject, request.Name)
	if err != nil {
		return failure(w, r, "Failed to create save", err)
	}

	return xhttp.WriteResponse(w, http.StatusCreated, newResponse(record))
}

func (h *Handlers) rename(w http.ResponseWriter, r *http.Request, session store.Session) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	}
	request, ok := decode(w, r)
	if !ok {
		return nil
	}

	record, err := h.service.Rename(r.Context(), id, session.Subject, request.Name)
	if err != nil {
		return failure(w, r, "Failed to rename save", err)
	}

	return xhttp.WriteResponse(w, http.StatusOK, newResponse(record))
}

func (h *Handlers) delete(w http.ResponseWriter, r *http.Request, session store.Session) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	}

	if err := h.service.Delete(r.Context(), id, session.Subject); err != nil {
		return failure(w, r, "Failed to delete save", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func decode(w http.ResponseWriter, r *http.Request) (Request, bool) {
	var request Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err == nil && decoder.Decode(&struct{}{}) != io.EOF {
		err = errors.New("unexpected data after the JSON object")
	}
	if err != nil {
		_ = problem.Write(w, http.StatusBadRequest, "The request body is not valid JSON for a save")
		return Request{}, false
	}

	return request, true
}

func failure(w http.ResponseWriter, r *http.Request, message string, err error) error {
	var invalid InvalidNameError
	switch {
	case errors.As(err, &invalid):
		return problem.Write(w, http.StatusUnprocessableEntity, invalid.Error())
	case errors.Is(err, ErrLimit):
		return problem.Write(w, http.StatusUnprocessableEntity, ErrLimit.Error())
	case errors.Is(err, ErrNotFound):
		return problem.Write(w, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrInUse):
		return problem.Write(w, http.StatusConflict, ErrInUse.Error())
	}

	slog.ErrorContext(r.Context(), message+": "+err.Error())
	return problem.Write(w, http.StatusInternalServerError, "Something went wrong, try again later")
}
