package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceIncomes interface {
}

type IncomesHandler struct {
	service ServiceIncomes
}

func NewIncomesHandler(s ServiceIncomes) *IncomesHandler {
	return &IncomesHandler{s}
}

func (h *IncomesHandler) RegisterRoutes(r chi.Router) {
}

func (h *IncomesHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *IncomesHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *IncomesHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *IncomesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
