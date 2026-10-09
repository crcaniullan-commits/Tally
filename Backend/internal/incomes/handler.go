package incomes

import (
	"context"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceIncomes interface {
	AddIncome(context.Context, IncomePayload, uuid.UUID) (model.Income, error)
	DeleteIncome(context.Context, uuid.UUID, uuid.UUID) error
	GetAllIncomesOfUser(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Income, error)
}

type IncomesHandler struct {
	service ServiceIncomes
	errors  errorhandler.ErrorsResponse
}

func NewIncomesHandler(s ServiceIncomes, e errorhandler.ErrorsResponse) *IncomesHandler {
	return &IncomesHandler{s, e}
}

type IncomePayload struct {
	Monto         int64               `json:"monto" validate:"required,gt=0"`
	PaymentMethod model.PaymentMethod `json:"payment_method" validate:"required,oneof=debito credito transferencia efectivo"`
	Descripcion   *string             `json:"descripcion" validate:"omitempty,max=100"`
	CategoryID    *uuid.UUID          `json:"category_id" validate:"omitempty,uuid4"`
}

/*
*	Añade un nuveo income
 */
func (h *IncomesHandler) AddIncome(w http.ResponseWriter, r *http.Request) {
	var payload IncomePayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	user := util.GetUserFromContext(r)

	income, err := h.service.AddIncome(r.Context(), payload, user.ID)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusCreated, income); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}

/*
*	Elimina un income
 */
func (h *IncomesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	incomeID, err := uuid.Parse(chi.URLParam(r, "incomeID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	user := util.GetUserFromContext(r)

	if err = h.service.DeleteIncome(r.Context(), incomeID, user.ID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusOK, "income eliminado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

/*
*	Optiene todos los income del usuario
*
*	Acepta los query params de paginación: limit (1..30, default 30), offset y
*	el rango de fechas since/until (AAAA-MM-DD). Un parámetro mal formado es un
*	400, no se cae al default en silencio.
 */
func (h *IncomesHandler) GetIncomesOfUser(w http.ResponseWriter, r *http.Request) {
	user := util.GetUserFromContext(r)

	fq, err := pagination.NewIncomePaginationQuery().Parse(r)

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	incomes, err := h.service.GetAllIncomesOfUser(r.Context(), user.ID, fq)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, incomes); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}
