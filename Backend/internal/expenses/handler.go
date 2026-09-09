package expenses

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
)

type ServiceExpenses interface {
}

type ExpensesHandler struct {
	service ServiceExpenses
	error   errorhandler.ErrorsResponse
}

func NewExpensesHandler(s ServiceExpenses, e errorhandler.ErrorsResponse) *ExpensesHandler {
	return &ExpensesHandler{s, e}
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
