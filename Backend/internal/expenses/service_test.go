package expenses

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func newService(store StoreExpenses) *ExpenseService {
	return NewExpenseService(store)
}

func TestExpenseService_AddExpense(t *testing.T) {
	t.Run("crea el gasto con el category_id validado", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()
		expense := newTestExpense()

		store.On("CategoryExists", mock.Anything, testCategoryID, testEmpresaID).Return(true, nil).Once()
		store.On("AddExpense", mock.Anything, mock.MatchedBy(func(e *model.Expense) bool {
			return e.Monto == payload.Monto && e.EmpresaID == testEmpresaID && e.CategoryID == testCategoryID && e.Fecha.IsZero()
		})).Return(nil).Run(func(args mock.Arguments) {
			e := args.Get(1).(*model.Expense)
			e.ID = expense.ID
			e.Fecha = expense.Fecha
			e.CreatedAt = expense.CreatedAt
		}).Once()

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, expense.ID, result.ID)
		assert.Equal(t, testCategoryID, result.CategoryID)
		store.AssertExpectations(t)
	})

	t.Run("devuelve ErrNotFound si la categoria no existe", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoryExists", mock.Anything, testCategoryID, testEmpresaID).Return(false, nil).Once()

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("rechaza una categoria de otra empresa", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoryExists", mock.Anything, testCategoryID, testEmpresaID).Return(false, nil).Once()

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("devuelve ErrNotFound si no se provee categoria", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := ExpensePayload{Monto: 1500}

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "CategoryExists", mock.Anything, mock.Anything, mock.Anything)
		store.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything)
	})

	t.Run("propaga el error del store de categorias", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoryExists", mock.Anything, testCategoryID, testEmpresaID).Return(false, errStoreBoom).Once()

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		payload := ExpensePayload{Monto: 1500, CategoryID: &testCategoryID}

		store.On("CategoryExists", mock.Anything, testCategoryID, testEmpresaID).Return(true, nil).Once()
		store.On("AddExpense", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.AddExpense(context.Background(), payload, testEmpresaID)

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result.ID)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestExpenseService_DeleteExpense(t *testing.T) {
	t.Run("elimina el gasto correctamente", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		store.On("DeleteExpense", mock.Anything, testExpenseID, testEmpresaID).Return(nil).Once()

		err := service.DeleteExpense(context.Background(), testExpenseID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		store.On("DeleteExpense", mock.Anything, testExpenseID, testEmpresaID).Return(util.ErrNotFound).Once()

		err := service.DeleteExpense(context.Background(), testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})
}

func TestExpenseService_GetAllExpensesOfEmpresa(t *testing.T) {
	t.Run("devuelve los gastos de la empresa", func(t *testing.T) {
		store := new(StoreExpensesMock)
		service := newService(store)

		expenses := []model.Expense{newTestExpense()}
		fq := newTestFilterQuery()

		store.On("GetAllExpensesOfEmpresa", mock.Anything, testEmpresaID, fq).Return(expenses, nil).Once()

		result, err := service.GetAllExpensesOfEmpresa(context.Background(), testEmpresaID, fq)

		require.NoError(t, err)
		assert.Equal(t, expenses, result)
		store.AssertExpectations(t)
	})
}
