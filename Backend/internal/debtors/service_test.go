package debtors

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestDebtorService_Create(t *testing.T) {
	t.Run("crea la deuda con el payload y la empresa", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		payload := newTestDebtorPayload()

		store.On("Crear", mock.Anything, mock.MatchedBy(func(debtor *model.Debtor) bool {
			return debtor.EmpresaID == testEmpresaID &&
				debtor.Nombre == payload.Nombre &&
				debtor.Monto == payload.Monto &&
				debtor.Descripcion != nil &&
				*debtor.Descripcion == payload.Descripcion &&
				debtor.FechaLimite != nil &&
				debtor.FechaLimite.Equal(payload.FechaLimite)
		})).Return(nil).Run(func(args mock.Arguments) {
			debtor := args.Get(1).(*model.Debtor)
			debtor.ID = testDebtorID
			debtor.Pagado = false
			debtor.CreatedAt = testDebtorCreatedAt
		}).Once()

		debtor, err := service.Crear(context.Background(), testEmpresaID, payload)

		require.NoError(t, err)
		assert.Equal(t, testDebtorID, debtor.ID)
		assert.Equal(t, testEmpresaID, debtor.EmpresaID)
		assert.Equal(t, payload.Nombre, debtor.Nombre)
		assert.Equal(t, payload.Monto, debtor.Monto)
		require.NotNil(t, debtor.Descripcion)
		assert.Equal(t, payload.Descripcion, *debtor.Descripcion)
		require.NotNil(t, debtor.FechaLimite)
		assert.True(t, debtor.FechaLimite.Equal(payload.FechaLimite))
		assert.False(t, debtor.Pagado)
		assert.Equal(t, testDebtorCreatedAt, debtor.CreatedAt)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		store.On("Crear", mock.Anything, mock.AnythingOfType("*model.Debtor")).
			Return(errStoreBoom).Once()

		debtor, err := service.Crear(context.Background(), testEmpresaID, newTestDebtorPayload())

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, model.Debtor{}, debtor)
		store.AssertExpectations(t)
	})
}

func TestDebtorService_Update(t *testing.T) {
	t.Run("actualiza la deuda con el payload y la empresa", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		payload := newTestUpdatePayload()

		store.On("Actualizar", mock.Anything, testDebtorID, testEmpresaID, mock.MatchedBy(func(debtor model.Debtor) bool {
			return debtor.Nombre == payload.Nombre &&
				debtor.Monto == payload.Monto &&
				debtor.FechaLimite != nil &&
				debtor.FechaLimite.Equal(payload.FechaLimite)
		})).Return(nil).Once()

		err := service.Actualizar(context.Background(), testDebtorID, testEmpresaID, payload)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		store.On("Actualizar", mock.Anything, testDebtorID, testEmpresaID, mock.Anything).
			Return(errStoreBoom).Once()

		err := service.Actualizar(context.Background(), testDebtorID, testEmpresaID, newTestUpdatePayload())

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestDebtorService_PaidDebtor(t *testing.T) {
	t.Run("marca la deuda como pagada", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		store.On("MarcarDeudorPagado", mock.Anything, testDebtorID, testEmpresaID).Return(nil).Once()

		err := service.MarcarDeudorPagado(context.Background(), testDebtorID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		store.On("MarcarDeudorPagado", mock.Anything, testDebtorID, testEmpresaID).
			Return(errStoreBoom).Once()

		err := service.MarcarDeudorPagado(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestDebtorService_GetDebtors(t *testing.T) {
	t.Run("devuelve las deudas de la empresa", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)
		debtors := []model.Debtor{newTestDebtor()}

		store.On("ObtenerDeudores", mock.Anything, testEmpresaID).Return(debtors, nil).Once()

		result, err := service.ObtenerDeudores(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, debtors, result)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(AlmacenDeudoresMock)
		service := NuevoServicioDeudores(store)

		store.On("ObtenerDeudores", mock.Anything, testEmpresaID).
			Return(nil, errStoreBoom).Once()

		result, err := service.ObtenerDeudores(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, result)
		store.AssertExpectations(t)
	})
}
