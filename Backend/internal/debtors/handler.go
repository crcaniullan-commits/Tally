package debtors

import (
	"context"
	"errors"
	"net/http"
	"time"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceDebtors interface {
	Create(context.Context, uuid.UUID, DebtorPayload) (model.Debtor, error)
	Update(context.Context, uuid.UUID, uuid.UUID, UpdatePayload) error
	PaidDebtor(context.Context, uuid.UUID, uuid.UUID) error
	GetDebtors(context.Context, uuid.UUID) ([]model.Debtor, error)
}

type DebtorsHandler struct {
	service ServiceDebtors
	error   errorhandler.ErrorsResponse
}

func NewDebtorsHandler(s ServiceDebtors, e errorhandler.ErrorsResponse) *DebtorsHandler {
	return &DebtorsHandler{s, e}
}

type DebtorPayload struct {
	Nombre      string    `json:"nombre" validate:"required,max=30"`
	Monto       int64     `json:"monto" validate:"required,gt=0"`
	Descripcion string    `json:"descripcion" validate:"required,max=100"`
	FechaLimite time.Time `json:"fecha_limite" validate:"required"`
}

func (h *DebtorsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload DebtorPayload

	if err := util.ReadJSONAndValidate(w, r, &payload); err != nil {
		h.error.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	deb, err := h.service.Create(r.Context(), empresaID, payload)

	if err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusCreated, deb); err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}
}

type UpdatePayload struct {
	Nombre      string    `json:"nombre" validate:"required,max=30"`
	Monto       int64     `json:"monto" validate:"required,gt=0"`
	FechaLimite time.Time `json:"fecha_limite" validate:"required"`
}

func (h *DebtorsHandler) Update(w http.ResponseWriter, r *http.Request) {
	debtorID, err := uuid.Parse(chi.URLParam(r, "debtorID"))

	if err != nil {
		h.error.BadRequestResponse(w, r, err)
		return
	}

	var payload UpdatePayload

	if err := util.ReadJSONAndValidate(w, r, &payload); err != nil {
		h.error.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	if err := h.service.Update(r.Context(), debtorID, empresaID, payload); err != nil {
		switch {
		case errors.Is(err, util.ErrNotFound):
			h.error.NotFoundResponse(w, r, err)
		default:
			h.error.InternalServerError(w, r, err)
		}
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, "Deuda actualizada"); err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}
}

func (h *DebtorsHandler) PaidDebtor(w http.ResponseWriter, r *http.Request) {
	debtorID, err := uuid.Parse(chi.URLParam(r, "debtorID"))

	if err != nil {
		h.error.BadRequestResponse(w, r, err)
		return
	}

	empresaID := util.GetEmpresaFromContext(r).ID

	if err := h.service.PaidDebtor(r.Context(), debtorID, empresaID); err != nil {
		switch {
		case errors.Is(err, util.ErrNotFound):
			h.error.NotFoundResponse(w, r, err)
		default:
			h.error.InternalServerError(w, r, err)
		}
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, "Deuda pagada"); err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}
}

func (h *DebtorsHandler) GetDebtorsOfEmpresa(w http.ResponseWriter, r *http.Request) {
	empresaID := util.GetEmpresaFromContext(r).ID

	debtors, err := h.service.GetDebtors(r.Context(), empresaID)
	if err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, debtors); err != nil {
		h.error.InternalServerError(w, r, err)
		return
	}
}
