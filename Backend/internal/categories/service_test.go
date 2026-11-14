package categories

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

func newService(store RepositorioCategorias) *ServicioCategorias {
	return NuevoServicioCategorias(store)
}

func TestNormalizarNombre(t *testing.T) {
	casos := []struct {
		nombre   string
		esperado string
	}{
		{"Comida", "comida"},
		{"comida", "comida"},
		{"COMIDA", "comida"},
		{"Áqui", "aqui"},
		{"aqui", "aqui"},
		{"áqui", "aqui"},
		{"Café", "cafe"},
		{"CAFÉ", "cafe"},
		{"  Comida  ", "comida"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			assert.Equal(t, caso.esperado, normalizarNombre(caso.nombre))
		})
	}
}

func TestCategorieService_Create(t *testing.T) {
	t.Run("crea la categoria correctamente", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID
		categoria := newTestCategoria()

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, empresaID).Return(nil, nil).Once()

		store.On("Crear", mock.Anything, mock.MatchedBy(func(c *model.Categorie) bool {
			return c.Nombre == payload.Nombre && c.EmpresaID != nil && *c.EmpresaID == empresaID
		})).Return(nil).Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.Categorie)
			c.ID = categoria.ID
		}).Once()

		result, err := service.Crear(context.Background(), payload, &empresaID)

		require.NoError(t, err)
		assert.Equal(t, categoria.ID, result.ID)
		assert.Equal(t, categoria.Nombre, result.Nombre)
		store.AssertExpectations(t)
	})

	t.Run("no crea la categoria si ya existe el nombre en la misma empresa", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, empresaID).
			Return([]model.Categorie{newTestCategoria()}, nil).Once()

		result, err := service.Crear(context.Background(), payload, &empresaID)

		require.ErrorIs(t, err, ErrorCategoriaDuplicada)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("considera duplicado el mismo nombre con otras mayusculas o acentos", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "CAFE"}
		empresaID := testEmpresaID

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, empresaID).
			Return([]model.Categorie{{ID: uuid.New(), Nombre: "Café", EmpresaID: &empresaID}}, nil).Once()

		result, err := service.Crear(context.Background(), payload, &empresaID)

		require.ErrorIs(t, err, ErrorCategoriaDuplicada)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error al buscar duplicados", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, empresaID).Return(nil, errStoreBoom).Once()

		result, err := service.Crear(context.Background(), payload, &empresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, uuid.Nil, result.ID)
		store.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store al crear", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := newTestPayloadCategoria()
		empresaID := testEmpresaID

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, empresaID).Return(nil, nil).Once()
		store.On("Crear", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		result, err := service.Crear(context.Background(), payload, &empresaID)

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

		payload := PayloadCategoria{Nombre: "Bebidas"}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).
			Return([]model.Categorie{newTestCategoria()}, nil).Once()
		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(nil).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("no cambia el nombre si otra categoria de la empresa ya lo usa", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "Bebidas"}
		existente := model.Categorie{ID: uuid.New(), Nombre: "bebidas", EmpresaID: &testEmpresaID}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).
			Return([]model.Categorie{existente}, nil).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, ErrorCategoriaDuplicada)
		store.AssertNotCalled(t, "CambiarNombre", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("permite dejar el mismo nombre a la categoria que se renombra", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "COMIDA"}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).
			Return([]model.Categorie{newTestCategoria()}, nil).Once()
		store.On("CambiarNombre", mock.Anything, "COMIDA", testCategoriaID, testEmpresaID).Return(nil).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando la categoria no existe", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "Bebidas"}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).Return(nil, nil).Once()
		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error al buscar duplicados", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "Bebidas"}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).Return(nil, errStoreBoom).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertNotCalled(t, "CambiarNombre", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		payload := PayloadCategoria{Nombre: "Bebidas"}

		store.On("ObtenerCategoriasPorEmpresa", mock.Anything, testEmpresaID).Return(nil, nil).Once()
		store.On("CambiarNombre", mock.Anything, "Bebidas", testCategoriaID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.CambiarNombre(context.Background(), payload, testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestCategorieService_GetCategorias(t *testing.T) {
	t.Run("devuelve las categorias visibles para la empresa", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		global := model.Categorie{
			ID:     uuid.MustParse("c1d2e3f4-5a6b-4c7d-8e9f-0a1b2c3d4e5f"),
			Nombre: "Ventas",
		}
		esperadas := []model.Categorie{newTestCategoria(), global}

		store.On("ObtenerCategoriasVisibles", mock.Anything, testEmpresaID).
			Return(esperadas, nil).Once()

		categorias, err := service.ObtenerCategorias(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, esperadas, categorias)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("ObtenerCategoriasVisibles", mock.Anything, testEmpresaID).
			Return(nil, errStoreBoom).Once()

		categorias, err := service.ObtenerCategorias(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, categorias)
		store.AssertExpectations(t)
	})
}

func TestCategorieService_BorrarCategoria(t *testing.T) {
	t.Run("borra la categoria correctamente", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("TieneMovimientos", mock.Anything, testCategoriaID, testEmpresaID).Return(false, nil).Once()
		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(nil).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("no borra una categoria que tiene movimientos", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("TieneMovimientos", mock.Anything, testCategoriaID, testEmpresaID).Return(true, nil).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, ErrorCategoriaConMovimientos)
		store.AssertNotCalled(t, "BorrarCategoria", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga el error del chequeo de movimientos", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("TieneMovimientos", mock.Anything, testCategoriaID, testEmpresaID).Return(false, errStoreBoom).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertNotCalled(t, "BorrarCategoria", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando la categoria no existe", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("TieneMovimientos", mock.Anything, testCategoriaID, testEmpresaID).Return(false, nil).Once()
		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(CategoriaStoreMock)
		service := newService(store)

		store.On("TieneMovimientos", mock.Anything, testCategoriaID, testEmpresaID).Return(false, nil).Once()
		store.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}
