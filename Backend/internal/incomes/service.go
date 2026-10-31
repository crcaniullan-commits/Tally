package incomes

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreIncomes interface {
	AddIncome(context.Context, *model.Income) error
	DeleteIncome(context.Context, uuid.UUID, uuid.UUID) error
	GetAllIncomesOfEmpresa(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Income, error)
	CategoryExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
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

	// La categoría es opcional (la columna es nullable). Si viene, tiene que
	// existir y ser global o de la empresa; una categoría de otra empresa se
	// rechaza igual que una inexistente.
	if payload.CategoryID != nil {
		existe, err := s.store.CategoryExists(ctx, *payload.CategoryID, empresaID)
		if err != nil {
			return model.Income{}, err
		}
		if !existe {
			return model.Income{}, util.ErrNotFound
		}

		income.CategoryID = payload.CategoryID
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
