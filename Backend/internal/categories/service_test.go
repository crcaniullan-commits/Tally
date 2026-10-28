package categories

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

func newService(store CategoriaStore) *CategorieService {
	return NewCategorieService(store)
}

func TestCategorieService_Create(t *testing.T) {
	t.Run("crea la categoria correctamente", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID
		categoria := newTestCategoria()

		store.On("GetCategoriaPorNombreYEmpresa", mock.Anything, mock.MatchedBy(func(c model.Categorie) bool {
			return c.Nombre == payload.Nombre && c.EmpresaID != nil && *c.EmpresaID == empresaID
		})).Return(false, nil).Once()

		store.On("Create", mock.Anything, mock.MatchedBy(func(c *model.Categorie) bool {
			return c.Nombre == payload.Nombre && c.EmpresaID != nil && *c.EmpresaID == empresaID
		})).Return(nil).Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.Categorie)
			c.ID = categoria.ID
		}).Once()

		result, err := service.Create(context.Background(), payload, &empresaID)

		require.NoError(t, err)
		assert.Equal(t, categoria.ID, result.ID)
		assert.Equal(t, categoria.Nombre, result.Nombre)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store al crear", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID

		store.On("GetCategoriaPorNombreYEmpresa", mock.Anything, mock.Anything).Return(false, nil).Once()
		store.On("Create", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.Create(context.Background(), payload, &empresaID)

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result.ID)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestCategorieService_CambiarNombre(t *testing.T) {
	t.Run("cambia el nombre correctamente", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := CategoriaPayload{Nombre: "Bebidas"}

		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(nil).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound cuando la categoria no existe", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := CategoriaPayload{Nombre: "Bebidas"}

		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(util.ErrNotFound).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := CategoriaPayload{Nombre: "Bebidas"}

		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestCategorieService_BorrarCategoria(t *testing.T) {
	t.Run("borra la categoria correctamente", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(nil).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound cuando la categoria no existe", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(util.ErrNotFound).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}
