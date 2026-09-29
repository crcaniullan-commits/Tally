package incomes

import (
	"context"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type ServiceIncomes interface {
	AddIncome(context.Context, IncomePayload, uuid.UUID) error
	DeleteIncome(context.Context, uuid.UUID) error
	GetAllIncomesOfUser(context.Context, uuid.UUID) ([]IncomeStorage, error)
}

type IncomesHandler struct {
	service ServiceIncomes
	error   errorhandler.ErrorsResponse
}

func NewIncomesHandler(s ServiceIncomes, e errorhandler.ErrorsResponse) *IncomesHandler {
	return &IncomesHandler{s, e}
}

type IncomePayload struct {
	UserID        uuid.UUID          `json:"userID" validate:"required,uuid4"`
	Monto         float64            `json:"monto" validate:"required, gt=0"`
	PaymentMethod util.PaymentMethod `json:"payment_method" validate:"required, oneof=debito credito transferencia efectivo"`
	Descripcion   *string            `json:"description" validate:"omitempy,max=100"`
}

func (h *IncomesHandler) AddIncome(w http.ResponseWriter, r *http.Request) {
}

func (h *IncomesHandler) Delete(w http.ResponseWriter, r *http.Request) {
}

func (h *IncomesHandler) GetIncomesOfUser(w http.ResponseWriter, r *http.Request) {
}
