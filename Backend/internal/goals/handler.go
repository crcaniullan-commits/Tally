package goals

import (
	"context"
	"errors"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceGoals interface {
	createGoal(context.Context, CreateGoalPayload, uuid.UUID) (model.Goal, error)
	DeleteGoal(context.Context, uuid.UUID, uuid.UUID) error
	GetGoalState(context.Context, uuid.UUID, uuid.UUID) (GoalState, error)
	getGoal(context.Context, uuid.UUID, uuid.UUID) (model.Goal, error)
	getGoals(context.Context, uuid.UUID) ([]model.Goal, error)
}

type GoalsHandler struct {
	service ServiceGoals
	errors  errorhandler.ErrorsResponse
}

type GoalState struct {
	MontoActual        int64   `json:"monto_actual"`
	PorcentajeActual   float64 `json:"porcentaje_actual"`
	MontoPendiente     int64   `json:"monto_pendiente"`
	PorcentajeFaltante float64 `json:"porcentaje_faltante"`
}

type CreateGoalPayload struct {
	Nombre     string          `json:"nombre" validate:"required,max=30"`
	Periodo    util.GoalPeriod `json:"periodo" validate:"required,oneof=dia semana mes"`
	Monto_meta int64           `json:"monto_meta" validate:"required,gt=0"`
}

func NewGoalsHandler(s ServiceGoals, e errorhandler.ErrorsResponse) *GoalsHandler {
	return &GoalsHandler{s, e}
}

func (h *GoalsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload CreateGoalPayload
	if err := util.ReadJSONAndValidate(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r)

	if empresaID == nil {
		h.errors.BadRequestResponse(w, r, errors.New("not Empresa ID in the request"))
		return
	}

	goal, err := h.service.createGoal(r.Context(), payload, empresaID.ID)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusCreated, goal); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *GoalsHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := uuid.Parse(chi.URLParam(r, "goalID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	if err := h.service.DeleteGoal(r.Context(), goalID, empresaID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Objetivo eliminado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *GoalsHandler) GetGoalState(w http.ResponseWriter, r *http.Request) {
	goalID, err := uuid.Parse(chi.URLParam(r, "goalID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	state, err := h.service.GetGoalState(r.Context(), goalID, empresaID)

	if err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, state); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *GoalsHandler) getGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := uuid.Parse(chi.URLParam(r, "goalID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	goal, err := h.service.getGoal(r.Context(), goalID, empresaID)

	if err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, goal); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *GoalsHandler) getGoals(w http.ResponseWriter, r *http.Request) {

	empresaID := util.GetEmpresaFromContext(r).ID

	goals, err := h.service.getGoals(r.Context(), empresaID)

	if err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, goals); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
