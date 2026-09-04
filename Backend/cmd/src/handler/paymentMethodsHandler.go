package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServicePaymentMethods interface {
}

type PaymentMethodsHandler struct {
	service ServicePaymentMethods
}

func NewPaymentMethodsHandler(s ServicePaymentMethods) *PaymentMethodsHandler {
	return &PaymentMethodsHandler{s}
}

func (h *PaymentMethodsHandler) RegisterRoutes(r chi.Router) {
}

func (h *PaymentMethodsHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *PaymentMethodsHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *PaymentMethodsHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *PaymentMethodsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
