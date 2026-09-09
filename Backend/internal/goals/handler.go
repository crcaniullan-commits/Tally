package goals

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
)

type ServiceGoals interface {
}

type GoalsHandler struct {
	service ServiceGoals
	error   errorhandler.ErrorsResponse
}

func NewGoalsHandler(s ServiceGoals, e errorhandler.ErrorsResponse) *GoalsHandler {
	return &GoalsHandler{s, e}
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
