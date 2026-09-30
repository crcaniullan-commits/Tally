package incomes

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestIncomeService_AddIncome(t *testing.T) {
	t.Run("mapea el payload a IncomeStorage con el userID del usuario", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		payload := newTestPayload()

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *IncomeStorage) bool {
				return i.UserID == testUserID &&
					i.Monto == payload.Monto &&
					i.PaymentMethod == util.PaymentMethodDebito &&
					i.Descripcion != nil && *i.Descripcion == "Venta de almuerzo"
			}),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), payload, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertNumberOfCalls(t, "AddIncome", 1)
	})

	t.Run("acepta descripcion nula", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		payload := IncomePayload{Monto: 300, PaymentMethod: util.PaymentMethodEfectivo}

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *IncomeStorage) bool { return i.Descripcion == nil }),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), payload, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("no fija ID ni fechas: los deja para que los asigne la base", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *IncomeStorage) bool {
				return i.ID == uuid.Nil && i.Fecha.IsZero() && i.CreatedAt.IsZero()
			}),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("AddIncome", mock.Anything, mock.Anything).
			Return(errStoreBoom).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_DeleteIncome(t *testing.T) {
	t.Run("elimina el ingreso con el id recibido", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(nil).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "DeleteIncome", mock.Anything, testIncomeID, testUserID)
	})

	t.Run("propaga util.ErrNotFound cuando el ingreso no existe", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(util.ErrNotFound).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(errStoreBoom).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_GetAllIncomesOfUser(t *testing.T) {
	t.Run("devuelve la lista de ingresos del usuario", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		esperados := []IncomeStorage{newTestIncome()}
		store.On("GetAllIncomesOfUser", mock.Anything, testUserID).Return(esperados, nil).Once()

		incomes, err := service.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, esperados, incomes)
		store.AssertExpectations(t)
	})

	t.Run("devuelve slice nil sin error cuando el usuario no tiene ingresos", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfUser", mock.Anything, testUserID).Return(nil, nil).Once()

		incomes, err := service.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Nil(t, incomes, "sin ingresos debe quedar nil para que el JSON sea null")
		store.AssertExpectations(t)
	})

	t.Run("devuelve nil y el error si el store falla", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfUser", mock.Anything, testUserID).Return(nil, errStoreBoom).Once()

		incomes, err := service.GetAllIncomesOfUser(context.Background(), testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva incomes", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfUser", mock.Anything, testUserID).
			Return([]IncomeStorage{newTestIncome()}, errStoreBoom).Once()

		incomes, err := service.GetAllIncomesOfUser(context.Background(), testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes, "si hay error el servicio debe devolver nil")
		store.AssertExpectations(t)
	})

	t.Run("no filtra ingresos de otro usuario: consulta por el userID recibido", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		otroUsuario := uuid.New()
		store.On("GetAllIncomesOfUser", mock.Anything, otroUsuario).Return(nil, nil).Once()

		_, err := service.GetAllIncomesOfUser(context.Background(), otroUsuario)

		require.NoError(t, err)
		store.AssertCalled(t, "GetAllIncomesOfUser", mock.Anything, otroUsuario)
		store.AssertNumberOfCalls(t, "GetAllIncomesOfUser", 1)
	})
}
