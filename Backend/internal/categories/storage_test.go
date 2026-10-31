package categories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// mockStore envuelve un StoreCategorie sobre sqlmock y guarda el SQL que el
// código realmente mandó a la base.
type mockStore struct {
	store *StoreCategorie
	mock  sqlmock.Sqlmock
	sql   string
}

func newMockStore(t *testing.T) *mockStore {
	t.Helper()

	ms := &mockStore{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(_, actualSQL string) error {
			ms.sql = actualSQL
			return nil
		}),
	))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		mock.ExpectClose()
		_ = db.Close()
	})

	ms.store = NewStoreCategorie(db)
	ms.mock = mock

	return ms
}

func categoriaColumns() []string {
	return []string{"id", "nombre", "empresa_id"}
}

func categoriaRow() []driver.Value {
	return []driver.Value{
		testCategoriaID,
		"Comida",
		testEmpresaID,
	}
}

func TestStoreCategorie_GetCategoriasPorEmpresa(t *testing.T) {
	t.Run("devuelve las categorias de la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		bebidas := []driver.Value{
			uuid.MustParse("b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5e"),
			"Bebidas",
			testEmpresaID,
		}

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(categoriaRow()...).
				AddRow(bebidas...))

		categorias, err := ms.store.GetCategoriasPorEmpresa(context.Background(), testEmpresaID)

		require.NoError(t, err)
		require.Len(t, categorias, 2)
		assert.Equal(t, testCategoriaID, categorias[0].ID)
		assert.Equal(t, "Comida", categorias[0].Nombre)
		require.NotNil(t, categorias[0].EmpresaID)
		assert.Equal(t, testEmpresaID, *categorias[0].EmpresaID)
		assert.Equal(t, "Bebidas", categorias[1].Nombre)
		assert.Contains(t, ms.sql, "SELECT id, nombre, empresa_id")
		assert.Contains(t, ms.sql, "FROM categories")
		assert.Contains(t, ms.sql, "WHERE empresa_id = $1")
	})

	t.Run("devuelve una lista vacia si la empresa no tiene categorias", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()))

		categorias, err := ms.store.GetCategoriasPorEmpresa(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.Empty(t, categorias)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnError(errDBBoom)

		categorias, err := ms.store.GetCategoriasPorEmpresa(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, categorias)
	})

	t.Run("propaga el error al escanear una fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(123, "Comida", testEmpresaID))

		categorias, err := ms.store.GetCategoriasPorEmpresa(context.Background(), testEmpresaID)

		require.Error(t, err)
		assert.Nil(t, categorias)
	})

	t.Run("propaga el error de rows", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(categoriaRow()...).
				RowError(0, errDBBoom))

		categorias, err := ms.store.GetCategoriasPorEmpresa(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, categorias)
	})
}

func TestStoreCategorie_GetCategoriasVisibles(t *testing.T) {
	t.Run("devuelve las categorias de la empresa y las globales", func(t *testing.T) {
		ms := newMockStore(t)

		global := []driver.Value{
			uuid.MustParse("c1d2e3f4-5a6b-4c7d-8e9f-0a1b2c3d4e5f"),
			"Ventas",
			nil,
		}

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(categoriaRow()...).
				AddRow(global...))

		categorias, err := ms.store.GetCategoriasVisibles(context.Background(), testEmpresaID)

		require.NoError(t, err)
		require.Len(t, categorias, 2)
		require.NotNil(t, categorias[0].EmpresaID)
		assert.Equal(t, testEmpresaID, *categorias[0].EmpresaID)
		assert.Equal(t, "Ventas", categorias[1].Nombre)
		assert.Nil(t, categorias[1].EmpresaID, "las globales tienen empresa_id NULL")
		assert.Contains(t, ms.sql, "WHERE empresa_id IS NULL OR empresa_id = $1")
		assert.Contains(t, ms.sql, "ORDER BY empresa_id NULLS LAST, nombre")
	})

	t.Run("devuelve una lista vacia si no hay categorias visibles", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()))

		categorias, err := ms.store.GetCategoriasVisibles(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.Empty(t, categorias)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnError(errDBBoom)

		categorias, err := ms.store.GetCategoriasVisibles(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, categorias)
	})

	t.Run("propaga el error al escanear una fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(123, "Comida", testEmpresaID))

		categorias, err := ms.store.GetCategoriasVisibles(context.Background(), testEmpresaID)

		require.Error(t, err)
		assert.Nil(t, categorias)
	})

	t.Run("propaga el error de rows", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).
				AddRow(categoriaRow()...).
				RowError(0, errDBBoom))

		categorias, err := ms.store.GetCategoriasVisibles(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, categorias)
	})
}

func TestStoreCategorie_Create(t *testing.T) {
	t.Run("crea la categoria y devuelve el id", func(t *testing.T) {
		ms := newMockStore(t)

		empresaID := testEmpresaID
		categoria := &model.Categorie{
			Nombre:    "Comida",
			EmpresaID: &empresaID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs("Comida", &empresaID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testCategoriaID))

		err := ms.store.Create(context.Background(), categoria)

		require.NoError(t, err)
		assert.Equal(t, testCategoriaID, categoria.ID)
		assert.Contains(t, ms.sql, "INSERT INTO categories")
	})

	t.Run("propaga util.ErrNotFound si no retorna filas", func(t *testing.T) {
		ms := newMockStore(t)

		empresaID := testEmpresaID
		categoria := &model.Categorie{
			Nombre:    "Comida",
			EmpresaID: &empresaID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs("Comida", &empresaID).
			WillReturnError(sql.ErrNoRows)

		err := ms.store.Create(context.Background(), categoria)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		empresaID := testEmpresaID
		categoria := &model.Categorie{
			Nombre:    "Comida",
			EmpresaID: &empresaID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs("Comida", &empresaID).
			WillReturnError(errDBBoom)

		err := ms.store.Create(context.Background(), categoria)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreCategorie_CambiarNombre(t *testing.T) {
	t.Run("actualiza el nombre correctamente", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Bebidas", testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.CambiarNombre(context.Background(), "Bebidas", testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "UPDATE categories")
		assert.Contains(t, ms.sql, "SET nombre = $1")
		assert.Contains(t, ms.sql, "WHERE id = $2 and empresa_id = $3")
	})

	t.Run("devuelve util.ErrNotFound si no se actualizo ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Bebidas", testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.CambiarNombre(context.Background(), "Bebidas", testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Bebidas", testCategoriaID, testEmpresaID).
			WillReturnError(errDBBoom)

		err := ms.store.CambiarNombre(context.Background(), "Bebidas", testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Bebidas", testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.CambiarNombre(context.Background(), "Bebidas", testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreCategorie_BorrarCategoria(t *testing.T) {
	t.Run("elimina la categoria correctamente", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "DELETE FROM categories")
		assert.Contains(t, ms.sql, "WHERE id = $1 and empresa_id = $2")
	})

	t.Run("devuelve util.ErrNotFound si no se elimino ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testCategoriaID, testEmpresaID).
			WillReturnError(errDBBoom)

		err := ms.store.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testCategoriaID, testEmpresaID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.BorrarCategoria(context.Background(), testCategoriaID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}
