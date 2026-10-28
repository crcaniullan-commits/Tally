package empresas

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/crcaniullan-commits/Tally/internal/model"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func newService(store StoreEmpresas) *EmpresaService {
	return NewEmpresaService(store)
}

func TestEmpresaService_GetEmpresas(t *testing.T) {
	t.Run("devuelve las empresas del usuario", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		empresas := []model.Empresa{newTestEmpresa()}
		store.On("GetEmpresas", mock.Anything, testUserID).Return(empresas, nil).Once()

		result, err := service.GetEmpresas(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, empresas, result)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("GetEmpresas", mock.Anything, testUserID).Return(nil, errStoreBoom).Once()

		result, err := service.GetEmpresas(context.Background(), testUserID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestEmpresaService_Create(t *testing.T) {
	t.Run("crea la empresa correctamente", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		payload := newTestPayloadEmpresa()
		empresa := newTestEmpresa()

		store.On("Create", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
			emp, ok := e.(*model.Empresa)
			if !ok {
				return false
			}
			return emp.NombreEmpresa == payload.Nombre_empresa && emp.UserID == payload.user_id
		})).Return(nil).Run(func(args mock.Arguments) {
			emp := args.Get(1).(*model.Empresa)
			emp.ID = empresa.ID
		}).Once()

		result, err := service.Create(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, empresa.ID, result.ID)
		assert.Equal(t, empresa.NombreEmpresa, result.NombreEmpresa)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		payload := newTestPayloadEmpresa()

		store.On("Create", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.Create(context.Background(), payload)

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result.ID)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestEmpresaService_CambiarNombre(t *testing.T) {
	t.Run("cambia el nombre correctamente", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(nil).Once()

		err := service.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound cuando la empresa no existe", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(util.ErrNotFound).Once()

		err := service.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(errStoreBoom).Once()

		err := service.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestEmpresaService_BorrarEmpresa(t *testing.T) {
	t.Run("borra la empresa correctamente", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(nil).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound cuando la empresa no existe", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(util.ErrNotFound).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(StoreEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(errStoreBoom).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}
