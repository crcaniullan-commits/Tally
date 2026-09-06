package handler

import (
	"context"
	"net/http"
	"uuid"

	"github.com/crcaniullan-commits/Tally/cmd/src/httputil"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/store"
	"github.com/go-chi/chi/v5"
)

type ServiceUsers interface {
	Create(http.ResponseWriter, *http.Request) (*store.Users, error)
	Update(context.Context, uuid.UUID) error
	Delete(context.Context, uuid.UUID) error
	GetByID(context.Context, uuid.UUID) (*store.Users, error)
}

type UsersHandler struct {
	service ServiceUsers
	errors  errorhandler.ErrorsResponse
}

func NewUserHandler(s ServiceUsers, e errorhandler.ErrorsResponse) *UsersHandler {
	return &UsersHandler{s, e}
}

func (h *UsersHandler) RegisterRoutes(r chi.Router) chi.Router {
	return r.Route("/users", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", h.GetByID)
			r.Patch("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})
}

func (h *UsersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if err := httputil.JsonResponse(w, http.StatusOK, "funcionando"); err != nil {
		httputil.WriteJSONError(w, http.StatusInternalServerError, "error en ejecución")
		return
	}
}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {

	user, err := h.service.Create(w, r)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := httputil.JsonResponse(w, http.StatusCreated, user); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
}
