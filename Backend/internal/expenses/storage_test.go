package expenses

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// mockStore envuelve un StoreExpense sobre sqlmock y guarda el SQL.
type mockStore struct {
	store *StoreExpense
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

func expenseColumns() []string {
	return []string{"id", "empresa_id", "monto", "category_id", "descripcion", "fecha", "created_at"}
}

func expenseRow() []driver.Value {
	descripcion := "Compra insumos"
	return []driver.Value{
		testExpenseID,
		testEmpresaID,
		int64(1500),
		testCategoryID,
		descripcion,
		testFecha,
		testFecha.Add(2 * time.Hour),
	}
}

func TestStoreExpense_DeleteExpense(t *testing.T) {
	t.Run("elimina el gasto filtrando por id y empresa_id", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testExpenseID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.DeleteExpense(context.Background(), testExpenseID, testEmpresaID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "DELETE FROM expenses")
		assert.Contains(t, ms.sql, "WHERE id = $1 and empresa_id = $2")
	})

	t.Run("devuelve util.ErrNotFound si no se elimino ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testExpenseID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.DeleteExpense(context.Background(), testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testExpenseID, testEmpresaID).
			WillReturnError(errDBBoom)

		err := ms.store.DeleteExpense(context.Background(), testExpenseID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreExpense_AddExpense(t *testing.T) {
	t.Run("inserta el gasto y retorna id, fecha, created_at", func(t *testing.T) {
		ms := newMockStore(t)

		descripcion := "Compra insumos"
		expense := &model.Expense{
			EmpresaID:   testEmpresaID,
			Monto:       1500,
			CategoryID:  testCategoryID,
			Descripcion: &descripcion,
			Fecha:       testFecha,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, int64(1500), testCategoryID, &descripcion, testFecha).
			WillReturnRows(sqlmock.NewRows([]string{"id", "fecha", "created_at"}).AddRow(testExpenseID, testFecha, testFecha.Add(2*time.Hour)))

		err := ms.store.AddExpense(context.Background(), expense)

		require.NoError(t, err)
		assert.Equal(t, testExpenseID, expense.ID)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		expense := &model.Expense{
			EmpresaID:  testEmpresaID,
			Monto:      1500,
			CategoryID: testCategoryID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, int64(1500), testCategoryID, (*string)(nil), time.Time{}).
			WillReturnError(errDBBoom)

		err := ms.store.AddExpense(context.Background(), expense)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreExpense_GetAllExpensesOfEmpresa(t *testing.T) {
	t.Run("mapea las columnas a model.Expense", func(t *testing.T) {
		ms := newMockStore(t)

		fq := pagination.NewIncomePaginationQuery()

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, fq.Limit, fq.Offset, fq.Since, fq.Until, fq.CategoryID).
			WillReturnRows(sqlmock.NewRows(expenseColumns()).AddRow(expenseRow()...))

		expenses, err := ms.store.GetAllExpensesOfEmpresa(context.Background(), testEmpresaID, fq)

		require.NoError(t, err)
		require.Len(t, expenses, 1)
		assert.Equal(t, testExpenseID, expenses[0].ID)
		assert.Equal(t, testEmpresaID, expenses[0].EmpresaID)
		assert.Equal(t, int64(1500), expenses[0].Monto)
		assert.Equal(t, testCategoryID, expenses[0].CategoryID)
		assert.Contains(t, ms.sql, "NULLIF($4, '')::date",
			"un since vacío no se puede castear a date")
		assert.Contains(t, ms.sql, "NULLIF($6, '')::uuid",
			"un category_id vacío no se puede castear a uuid")
		assert.Contains(t, ms.sql, "ORDER BY fecha DESC, created_at DESC, id DESC",
			"el id desempata para que la paginación no repita ni saltee filas")
	})
}

func TestStoreExpense_GetCategoryIDByName(t *testing.T) {
	columnas := []string{"id", "nombre"}

	t.Run("resuelve categoria por nombre incluyendo globales", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(columnas).AddRow(testCategoryID, testCategoryName))

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), testCategoryName, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testCategoryID, categoryID)
		assert.Contains(t, ms.sql, "(empresa_id IS NULL OR empresa_id = $1)",
			"tiene que aceptar las globales y las propias de la empresa")
	})

	t.Run("ignora mayúsculas y acentos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(columnas).AddRow(testCategoryID, "Àmígô"))

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), "amigo", testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testCategoryID, categoryID)
	})

	t.Run("prefiere la categoría de la empresa sobre la global", func(t *testing.T) {
		ms := newMockStore(t)

		globalID := uuid.MustParse("c1d2e3f4-5a6b-4c7d-8e9f-0a1b2c3d4e5f")

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(columnas).
				AddRow(testCategoryID, "Comida").
				AddRow(globalID, "comida"))

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), "COMIDA", testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testCategoryID, categoryID)
	})

	t.Run("devuelve ErrNotFound si no existe", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(columnas).AddRow(uuid.New(), "Bebidas"))

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), testCategoryName, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, categoryID)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnError(errDBBoom)

		_, err := ms.store.GetCategoryIDByName(context.Background(), testCategoryName, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}
