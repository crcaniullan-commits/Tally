package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceAccessKey interface {
}

type AccessKeysHandler struct {
	service ServiceAccessKey
}

func NewAccessKeysHandler(s ServiceAccessKey) *AccessKeysHandler {
	return &AccessKeysHandler{s}
}

func (h *AccessKeysHandler) RegisterRoutes(r chi.Router) {
}

func (h *AccessKeysHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *AccessKeysHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *AccessKeysHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *AccessKeysHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
