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
	GetCategoryIDByName(context.Context, string, uuid.UUID) (uuid.UUID, error)
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
		Fecha:       payload.Fecha,
	}

	if payload.CategoryID != nil {
		expense.CategoryID = *payload.CategoryID
	} else if payload.Categoria != nil && *payload.Categoria != "" {
		categoryID, err := s.store.GetCategoryIDByName(ctx, *payload.Categoria, empresaID)
		if err != nil {
			if err == util.ErrNotFound {
				return model.Expense{}, util.ErrNotFound
			}
			return model.Expense{}, err
		}
		expense.CategoryID = categoryID
	} else {
		return model.Expense{}, util.ErrNotFound
	}

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
