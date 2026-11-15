package expenses

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func newService(store RepositorioGastos) *ServicioGastos {
	return NuevoServicioGastos(store)
}

func TestExpenseService_AddExpense(t *testing.T) {
	t.Run("crea el gasto con el category_id validado", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()
		expense := newTestExpense()

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID, model.CategoryTypeGasto).Return(true, nil).Once()
		store.On("AgregarGasto", mock.Anything, mock.MatchedBy(func(e *model.Expense) bool {
			return e.Monto == payload.Monto && e.EmpresaID == testEmpresaID && e.CategoryID == testCategoryID && e.Fecha.IsZero() &&
				e.PaymentMethod == model.PaymentMethodDebito
		})).Return(nil).Run(func(args mock.Arguments) {
			e := args.Get(1).(*model.Expense)
			e.ID = expense.ID
			e.Fecha = expense.Fecha
			e.CreatedAt = expense.CreatedAt
		}).Once()

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, expense.ID, result.ID)
		assert.Equal(t, testCategoryID, result.CategoryID)
		store.AssertExpectations(t)
	})

	t.Run("devuelve ErrNotFound si la categoria no existe", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID, model.CategoryTypeGasto).Return(false, nil).Once()

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("rechaza una categoria de otra empresa", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID, model.CategoryTypeGasto).Return(false, nil).Once()

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("devuelve ErrNotFound si no se provee categoria", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := PayloadGasto{Monto: 1500}

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "CategoriaExiste", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		store.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything)
	})

	t.Run("propaga el error del store de categorias", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := newTestPayloadExpense()

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID, model.CategoryTypeGasto).Return(false, errStoreBoom).Once()

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		payload := PayloadGasto{Monto: 1500, CategoryID: &testCategoryID}

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID, model.CategoryTypeGasto).Return(true, nil).Once()
		store.On("AgregarGasto", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.AgregarGasto(context.Background(), payload, testEmpresaID)

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result.ID)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestExpenseService_DeleteExpense(t *testing.T) {
	t.Run("elimina el gasto correctamente", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		store.On("EliminarGasto", mock.Anything, testExpenseID, testEmpresaID).Return(nil).Once()

		err := service.EliminarGasto(context.Background(), testExpenseID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		store.On("EliminarGasto", mock.Anything, testExpenseID, testEmpresaID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.EliminarGasto(context.Background(), testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})
}

func TestExpenseService_GetAllExpensesOfEmpresa(t *testing.T) {
	t.Run("devuelve los gastos de la empresa", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		expenses := []model.Expense{newTestExpense()}
		fq := newTestFilterQuery()

		store.On("ObtenerGastosDeEmpresa", mock.Anything, testEmpresaID, fq).Return(expenses, nil).Once()

		result, err := service.ObtenerGastosDeEmpresa(context.Background(), testEmpresaID, fq)

		require.NoError(t, err)
		assert.Equal(t, expenses, result)
		store.AssertExpectations(t)
	})
}

func TestExpenseService_UpdateExpense(t *testing.T) {
	t.Run("manda al store solo los campos del payload", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		monto := int64(2500)
		descripcion := "Compra de insumos"

		store.On("ActualizarGasto", mock.Anything,
			mock.MatchedBy(func(c CambiosGasto) bool {
				return c.Monto != nil && *c.Monto == monto &&
					c.CategoryID == nil &&
					c.Descripcion != nil && *c.Descripcion == descripcion
			}),
			testExpenseID, testEmpresaID,
		).Return(newTestExpense(), nil).Once()

		expense, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			Monto:       &monto,
			Descripcion: &descripcion,
		}, testExpenseID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testExpenseID, expense.ID)
		store.AssertExpectations(t)
	})

	t.Run("no manda la fecha: el movimiento conserva la suya", func(t *testing.T) {
		// La fecha no es un campo editable: el store solo recibe los campos
		// que el service decide tocar.
		store := new(RepositorioExpensesMock)
		service := newService(store)

		monto := int64(2500)

		store.On("ActualizarGasto", mock.Anything,
			mock.MatchedBy(func(c CambiosGasto) bool {
				return c.Monto != nil && *c.Monto == monto
			}),
			testExpenseID, testEmpresaID,
		).Return(newTestExpense(), nil).Once()

		_, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			Monto: &monto,
		}, testExpenseID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("cambia el metodo de pago", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		metodo := model.PaymentMethodTransferencia

		store.On("ActualizarGasto", mock.Anything,
			mock.MatchedBy(func(c CambiosGasto) bool {
				return c.PaymentMethod != nil && *c.PaymentMethod == model.PaymentMethodTransferencia
			}),
			testExpenseID, testEmpresaID,
		).Return(newTestExpense(), nil).Once()

		expense, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			PaymentMethod: &metodo,
		}, testExpenseID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testExpenseID, expense.ID)
		store.AssertExpectations(t)
	})

	t.Run("rechaza un payload sin campos", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		_, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{}, testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, ErrorSinCambios)
		store.AssertNotCalled(t, "ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("valida la categoria antes de actualizar", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		categoria := uuid.New()

		store.On("CategoriaExiste", mock.Anything, categoria, testEmpresaID, model.CategoryTypeGasto).Return(false, nil).Once()

		_, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			CategoryID: &categoria,
		}, testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertNotCalled(t, "ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("traduce el gasto inexistente a ErrorMovimientoNoEncontrado", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		monto := int64(2500)

		store.On("ActualizarGasto", mock.Anything, mock.Anything, testExpenseID, testEmpresaID).
			Return(model.Expense{}, utilidades.ErrorNoEncontrado).Once()

		_, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			Monto: &monto,
		}, testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, ErrorMovimientoNoEncontrado)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioExpensesMock)
		service := newService(store)

		monto := int64(2500)

		store.On("ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.Expense{}, errStoreBoom).Once()

		_, err := service.ActualizarGasto(context.Background(), PayloadActualizarGasto{
			Monto: &monto,
		}, testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
	})
}
