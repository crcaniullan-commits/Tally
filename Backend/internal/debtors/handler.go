package debtors

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
)

type ServiceDebtors interface {
}

type DebtorsHandler struct {
	service ServiceDebtors
	error   errorhandler.ErrorsResponse
}

func NewDebtorsHandler(s ServiceDebtors, e errorhandler.ErrorsResponse) *DebtorsHandler {
	return &DebtorsHandler{s, e}
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
