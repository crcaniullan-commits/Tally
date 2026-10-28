package expenses

import (
	"context"
	"net/http"
	"time"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceExpenses interface {
	AddExpense(context.Context, ExpensePayload, uuid.UUID) (model.Expense, error)
	DeleteExpense(context.Context, uuid.UUID, uuid.UUID) error
	GetAllExpensesOfEmpresa(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Expense, error)
}

type ExpensesHandler struct {
	service ServiceExpenses
	errors  errorhandler.ErrorsResponse
}

func NewExpensesHandler(s ServiceExpenses, e errorhandler.ErrorsResponse) *ExpensesHandler {
	return &ExpensesHandler{s, e}
}

func empresaID(r *http.Request) uuid.UUID {
	return util.GetEmpresaFromContext(r).ID
}

type ExpensePayload struct {
	Monto       int64      `json:"monto" validate:"required,gt=0"`
	CategoryID  *uuid.UUID `json:"category_id" validate:"omitempty"`
	Categoria   *string    `json:"categoria" validate:"omitempty,max=100"`
	Descripcion *string    `json:"descripcion" validate:"omitempty,max=100"`
	Fecha       time.Time  `json:"fecha"`
}

func (h *ExpensesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload ExpensePayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	expense, err := h.service.AddExpense(r.Context(), payload, empresaID(r))

	if err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusCreated, expense); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *ExpensesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	expenseID, err := uuid.Parse(chi.URLParam(r, "expenseID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err = h.service.DeleteExpense(r.Context(), expenseID, empresaID(r)); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusOK, "expense eliminado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *ExpensesHandler) GetExpensesOfEmpresa(w http.ResponseWriter, r *http.Request) {
	fq, err := pagination.NewIncomePaginationQuery().Parse(r)

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	expenses, err := h.service.GetAllExpensesOfEmpresa(r.Context(), empresaID(r), fq)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, expenses); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
