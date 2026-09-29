package incomes

import (
	"context"

	"github.com/google/uuid"
)

type StoreIncomes interface {
	AddIncome(context.Context, IncomeStorage, uuid.UUID) error
	DeleteIncome(context.Context, uuid.UUID) error
	GetAllIncomesOfUser(context.Context, uuid.UUID) ([]IncomeStorage, error)
}
type IncomeService struct {
	store StoreIncome
}

func NewIncomeService(store StoreIncome) *IncomeService {
	return &IncomeService{store: store}
}

func (s *IncomeService) AddIncome(ctx context.Context, payload IncomePayload, userID uuid.UUID) error {
	income := &IncomeStorage{
		UserID:        payload.UserID,
		Monto:         payload.Monto,
		PaymentMethod: payload.PaymentMethod,
		Descripcion:   payload.Descripcion,
	}

	if err := s.store.AddIncome(ctx, *income, userID); err != nil {
		return err
	}

	return nil
}

func (s *IncomeService) DeleteIncome(ctx context.Context, incomeID uuid.UUID) error {
	if err := s.store.DeleteIncome(ctx, incomeID); err != nil {
		return err
	}
	return nil
}

func (s *IncomeService) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID) ([]IncomeStorage, error) {
	var incomes []IncomeStorage

	incomes, err := s.store.GetAllIncomesOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return incomes, nil
}
