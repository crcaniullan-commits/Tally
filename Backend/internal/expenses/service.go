package expenses

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreExpenses interface {
	AddExpense(context.Context, *model.Expense) error
	DeleteExpense(context.Context, uuid.UUID, uuid.UUID) error
	GetAllExpensesOfEmpresa(context.Context, uuid.UUID, pagination.IncomePaginationQuery) ([]model.Expense, error)
	CategoryExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}

type ExpenseService struct {
	store StoreExpenses
}

func NewExpenseService(store StoreExpenses) *ExpenseService {
	return &ExpenseService{store: store}
}

func (s *ExpenseService) AddExpense(ctx context.Context, payload ExpensePayload, empresaID uuid.UUID) (model.Expense, error) {
	expense := &model.Expense{
		EmpresaID:   empresaID,
		Monto:       payload.Monto,
		Descripcion: payload.Descripcion,
	}

	if payload.CategoryID == nil {
		return model.Expense{}, util.ErrNotFound
	}

	existe, err := s.store.CategoryExists(ctx, *payload.CategoryID, empresaID)
	if err != nil {
		return model.Expense{}, err
	}
	if !existe {
		return model.Expense{}, util.ErrNotFound
	}

	expense.CategoryID = *payload.CategoryID

	if err := s.store.AddExpense(ctx, expense); err != nil {
		return model.Expense{}, err
	}

	return *expense, nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, expenseID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.store.DeleteExpense(ctx, expenseID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *ExpenseService) GetAllExpensesOfEmpresa(ctx context.Context, empresaID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Expense, error) {
	expenses, err := s.store.GetAllExpensesOfEmpresa(ctx, empresaID, fq)

	if err != nil {
		return nil, err
	}

	return expenses, nil
}
