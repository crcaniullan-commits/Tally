package handler

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/cmd/src/httputil"
	"github.com/go-chi/chi/v5"
)

type ServiceUsers interface {
	Create()
	Update()
	Delete()
	GetByID()
}

type UsersHandler struct {
	service ServiceUsers
}

func NewUserHandler(s ServiceUsers) *UsersHandler {
	return &UsersHandler{s}
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
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
}
