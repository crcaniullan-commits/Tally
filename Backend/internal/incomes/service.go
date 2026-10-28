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
	GetAllIncomesOfEmpresa(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Income, error)
	GetCategoryIDByName(context.Context, string, uuid.UUID) (uuid.UUID, error)
}
type IncomeService struct {
	store StoreIncomes
}

func NewIncomeService(store StoreIncomes) *IncomeService {
	return &IncomeService{store}
}

func (s *IncomeService) AddIncome(ctx context.Context, payload IncomePayload, empresaID uuid.UUID) (model.Income, error) {
	income := &model.Income{
		EmpresaID:     empresaID,
		Monto:         payload.Monto,
		PaymentMethod: payload.PaymentMethod,
		Descripcion:   payload.Descripcion,
	}

	// La categoría llega por nombre porque el usuario no conoce su UUID. Si no
	// la manda, el ingreso queda sin categoría (la columna es nullable).
	if payload.Categoria != nil {
		categoryID, err := s.store.GetCategoryIDByName(ctx, *payload.Categoria, empresaID)
		if err != nil {
			return model.Income{}, err
		}

		income.CategoryID = &categoryID
	}

	if err := s.store.AddIncome(ctx, income); err != nil {
		return model.Income{}, err
	}

	return *income, nil
}

func (s *IncomeService) DeleteIncome(ctx context.Context, incomeID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.store.DeleteIncome(ctx, incomeID, empresaID); err != nil {
		return err
	}
	return nil
}

func (s *IncomeService) GetAllIncomesOfEmpresa(ctx context.Context, empresaID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Income, error) {
	incomes, err := s.store.GetAllIncomesOfEmpresa(ctx, empresaID, fq)
	if err != nil {
		return nil, err
	}

	return incomes, nil
}
