package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceExpenses interface {
}

type ExpensesHandler struct {
	service ServiceExpenses
}

func NewExpensesHandler(s ServiceExpenses) *ExpensesHandler {
	return &ExpensesHandler{s}
}

func (h *ExpensesHandler) RegisterRoutes(r chi.Router) {
}

func (h *ExpensesHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *ExpensesHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *ExpensesHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *ExpensesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
