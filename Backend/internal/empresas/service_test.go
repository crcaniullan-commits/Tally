package empresas

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/crcaniullan-commits/Tally/internal/model"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func newService(store RepositorioEmpresas) *ServicioEmpresas {
	return NuevoServicioEmpresas(store)
}

func TestEmpresaService_GetEmpresas(t *testing.T) {
	t.Run("devuelve las empresas del usuario", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		empresas := []model.Empresa{newTestEmpresa()}
		store.On("ObtenerEmpresas", mock.Anything, testUserID).Return(empresas, nil).Once()

		result, err := service.ObtenerEmpresas(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, empresas, result)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("ObtenerEmpresas", mock.Anything, testUserID).Return(nil, errStoreBoom).Once()

		result, err := service.ObtenerEmpresas(context.Background(), testUserID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestEmpresaService_Create(t *testing.T) {
	t.Run("crea la empresa correctamente", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		payload := newTestPayloadEmpresa()
		empresa := newTestEmpresa()

		store.On("Crear", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
			emp, ok := e.(*model.Empresa)
			if !ok {
				return false
			}
			return emp.NombreEmpresa == payload.NombreEmpresa && emp.UserID == payload.userID
		})).Return(nil).Run(func(args mock.Arguments) {
			emp := args.Get(1).(*model.Empresa)
			emp.ID = empresa.ID
		}).Once()

		result, err := service.Crear(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, empresa.ID, result.ID)
		assert.Equal(t, empresa.NombreEmpresa, result.NombreEmpresa)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		payload := newTestPayloadEmpresa()

		store.On("Crear", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.Crear(context.Background(), payload)

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result.ID)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestEmpresaService_CambiarNombre(t *testing.T) {
	t.Run("cambia el nombre correctamente", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(nil).Once()

		err := service.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando la empresa no existe", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
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
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(nil).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando la empresa no existe", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioEmpresasMock)
		service := newService(store)

		store.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(errStoreBoom).Once()

		err := service.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}
