package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceCategories interface {
}

type CategoriesHandler struct {
	service ServiceCategories
}

func NewCategoriesHandler(s ServiceCategories) *CategoriesHandler {
	return &CategoriesHandler{s}
}

func (h *CategoriesHandler) RegisterRoutes(r chi.Router) {
}

func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *CategoriesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
