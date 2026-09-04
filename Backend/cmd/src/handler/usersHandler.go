package handler

import (
	"net/http"

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

func (h *UsersHandler) RegisterRoutes(r chi.Router) {

}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
