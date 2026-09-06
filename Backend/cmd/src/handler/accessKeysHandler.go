package handler

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
)

type ServiceAccessKey interface {
}

type AccessKeysHandler struct {
	service ServiceAccessKey
	error   errorhandler.ErrorsResponse
}

func NewAccessKeysHandler(s ServiceAccessKey, e errorhandler.ErrorsResponse) *AccessKeysHandler {
	return &AccessKeysHandler{s, e}
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
