package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceGoals interface {
}

type GoalsHandler struct {
	service ServiceGoals
}

func NewGoalsHandler(s ServiceGoals) *GoalsHandler {
	return &GoalsHandler{s}
}

func (h *GoalsHandler) RegisterRoutes(r chi.Router) {
}

func (h *GoalsHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *GoalsHandler) Update(w http.ResponseWriter, r *http.Request) {
}

func (h *GoalsHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *GoalsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
}
