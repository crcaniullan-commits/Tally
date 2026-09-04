package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceDebtors interface {
}

type DebtorsHandler struct {
	service ServiceDebtors
}

func NewDebtorsHandler(s ServiceDebtors) *DebtorsHandler {
	return &DebtorsHandler{s}
}

func (h *DebtorsHandler) RegisterRoutes(r chi.Router) {
}

func (h *DebtorsHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *DebtorsHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *DebtorsHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *DebtorsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
