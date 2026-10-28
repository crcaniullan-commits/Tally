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

func TestStoreCategorie_GetCategoriaPorNombreYEmpresa(t *testing.T) {
	t.Run("devuelve false si no existe la categoria", func(t *testing.T) {
		ms := newMockStore(t)

		categoria := model.Categorie{Nombre: "Comida", EmpresaID: &testEmpresaID}
		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(&testEmpresaID, "Comida").
			WillReturnError(sql.ErrNoRows)

		exist, err := ms.store.GetCategoriaPorNombreYEmpresa(context.Background(), categoria)

		require.NoError(t, err)
		assert.False(t, exist)
		assert.Contains(t, ms.sql, "SELECT id, nombre, empresa_id")
		assert.Contains(t, ms.sql, "FROM categories")
		assert.Contains(t, ms.sql, "WHERE empresa_id = $1 and nombre = $2")
	})

	t.Run("devuelve true y AlreadyExistError si existe la categoria", func(t *testing.T) {
		ms := newMockStore(t)

		categoria := model.Categorie{Nombre: "Comida", EmpresaID: &testEmpresaID}
		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(&testEmpresaID, "Comida").
			WillReturnRows(sqlmock.NewRows(categoriaColumns()).AddRow(categoriaRow()...))

		exist, err := ms.store.GetCategoriaPorNombreYEmpresa(context.Background(), categoria)

		require.ErrorIs(t, err, AlreadyExistError)
		assert.True(t, exist)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		categoria := model.Categorie{Nombre: "Comida", EmpresaID: &testEmpresaID}
		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(&testEmpresaID, "Comida").
			WillReturnError(errDBBoom)

		exist, err := ms.store.GetCategoriaPorNombreYEmpresa(context.Background(), categoria)

		require.ErrorIs(t, err, errDBBoom)
		assert.True(t, exist)
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
		assert.Contains(t, ms.sql, "WHERE id = $2 and $3")
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
		assert.Contains(t, ms.sql, "WHERE id = $1 and $2")
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
