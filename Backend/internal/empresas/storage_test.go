package empresas

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

// mockStore envuelve un StoreEmpresa sobre sqlmock y guarda el SQL que el
// código realmente mandó a la base.
type mockStore struct {
	store *StoreEmpresa
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

	ms.store = NewStorage(db)
	ms.mock = mock

	return ms
}

func empresaColumns() []string {
	return []string{"id", "nombre_empresa", "user_id"}
}

func empresaRow() []driver.Value {
	return []driver.Value{
		testEmpresaID,
		"Mi Empresa",
		testUserID,
	}
}

func TestStoreEmpresa_GetByID(t *testing.T) {
	t.Run("obtiene la empresa por id y usuario", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testUserID).
			WillReturnRows(sqlmock.NewRows(empresaColumns()).AddRow(empresaRow()...))

		empresa, err := ms.store.GetByID(context.Background(), testEmpresaID, testUserID)

		require.NoError(t, err)
		require.NotNil(t, empresa)
		assert.Equal(t, testEmpresaID, empresa.ID)
		assert.Equal(t, "Mi Empresa", empresa.NombreEmpresa)
		assert.Equal(t, testUserID, empresa.UserID)
		assert.Contains(t, ms.sql, "SELECT * FROM empresas")
		assert.Contains(t, ms.sql, "WHERE id = $1 and user_id = $2")
	})

	t.Run("devuelve util.ErrNotFound si no encuentra la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testUserID).
			WillReturnError(sql.ErrNoRows)

		empresa, err := ms.store.GetByID(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
		require.Nil(t, empresa)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testUserID).
			WillReturnError(errDBBoom)

		empresa, err := ms.store.GetByID(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
		require.Nil(t, empresa)
	})
}

func TestStoreEmpresa_GetEmpresas(t *testing.T) {
	t.Run("devuelve todas las empresas del usuario", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(empresaColumns()).AddRow(empresaRow()...))

		empresas, err := ms.store.GetEmpresas(context.Background(), testUserID)

		require.NoError(t, err)
		require.Len(t, empresas, 1)
		assert.Equal(t, testEmpresaID, empresas[0].ID)
		assert.Equal(t, "Mi Empresa", empresas[0].NombreEmpresa)
		assert.Equal(t, testUserID, empresas[0].UserID)
		assert.Contains(t, ms.sql, "SELECT id, nombre_empresa, user_id")
		assert.Contains(t, ms.sql, "FROM empresas")
		assert.Contains(t, ms.sql, "WHERE user_id = $1")
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		empresas, err := ms.store.GetEmpresas(context.Background(), testUserID)

		require.ErrorIs(t, err, errDBBoom)
		require.Nil(t, empresas)
	})
}

func TestStoreEmpresa_Create(t *testing.T) {
	t.Run("crea la empresa y devuelve el id", func(t *testing.T) {
		ms := newMockStore(t)

		empresa := &model.Empresa{
			NombreEmpresa: "Mi Empresa",
			UserID:        testUserID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs("Mi Empresa", testUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testEmpresaID))

		err := ms.store.Create(context.Background(), empresa)

		require.NoError(t, err)
		assert.Equal(t, testEmpresaID, empresa.ID)
		assert.Contains(t, ms.sql, "INSERT INTO empresas")
		assert.Contains(t, ms.sql, "nombre_empresa")
		assert.Contains(t, ms.sql, "user_id")
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		empresa := &model.Empresa{
			NombreEmpresa: "Mi Empresa",
			UserID:        testUserID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs("Mi Empresa", testUserID).
			WillReturnError(errDBBoom)

		err := ms.store.Create(context.Background(), empresa)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreEmpresa_CambiarNombre(t *testing.T) {
	t.Run("actualiza el nombre correctamente", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Nueva Empresa", testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "UPDATE empresas")
		assert.Contains(t, ms.sql, "SET nombre_empresa = $1")
		assert.Contains(t, ms.sql, "WHERE id = $2 and user_id = $3")
	})

	t.Run("devuelve util.ErrNotFound si no se actualizo ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Nueva Empresa", testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Nueva Empresa", testEmpresaID, testUserID).
			WillReturnError(errDBBoom)

		err := ms.store.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs("Nueva Empresa", testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.CambiarNombre(context.Background(), "Nueva Empresa", testEmpresaID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreEmpresa_BorrarEmpresa(t *testing.T) {
	t.Run("elimina la empresa correctamente", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "DELETE FROM empresas")
		assert.Contains(t, ms.sql, "WHERE id = $1 and user_id = $2")
	})

	t.Run("devuelve util.ErrNotFound si no se elimino ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testEmpresaID, testUserID).
			WillReturnError(errDBBoom)

		err := ms.store.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testEmpresaID, testUserID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.BorrarEmpresa(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})
}
