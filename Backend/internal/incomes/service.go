package incomes

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/google/uuid"
)

type StoreIncomes interface {
	AddIncome(context.Context, *model.Income) error
	DeleteIncome(context.Context, uuid.UUID, uuid.UUID) error
	GetAllIncomesOfUser(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Income, error)
}
type IncomeService struct {
	store StoreIncomes
}

func NewIncomeService(store StoreIncomes) *IncomeService {
	return &IncomeService{store}
}

func (s *IncomeService) AddIncome(ctx context.Context, payload IncomePayload, userID uuid.UUID) (model.Income, error) {
	income := &model.Income{
		UserID:        userID,
		Monto:         payload.Monto,
		PaymentMethod: payload.PaymentMethod,
		Descripcion:   payload.Descripcion,
		CategoryID:    payload.CategoryID,
	}

	if err := s.store.AddIncome(ctx, income); err != nil {
		return model.Income{}, err
	}

	return *income, nil
}

func (s *IncomeService) DeleteIncome(ctx context.Context, incomeID uuid.UUID, userID uuid.UUID) error {
	if err := s.store.DeleteIncome(ctx, incomeID, userID); err != nil {
		return err
	}
	return nil
}

func (s *IncomeService) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Income, error) {
	incomes, err := s.store.GetAllIncomesOfUser(ctx, userID, fq)
	if err != nil {
		return nil, err
	}

	return incomes, nil
}
