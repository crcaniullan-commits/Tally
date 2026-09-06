package handler

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
)

type ServiceIncomes interface {
}

type IncomesHandler struct {
	service ServiceIncomes
	error   errorhandler.ErrorsResponse
}

func NewIncomesHandler(s ServiceIncomes, e errorhandler.ErrorsResponse) *IncomesHandler {
	return &IncomesHandler{s, e}
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
