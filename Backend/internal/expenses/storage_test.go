package expenses

import (
	"context"
	"database/sql"
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
	})
}

func TestStoreExpense_GetCategoryIDByName(t *testing.T) {
	t.Run("resuelve categoria por nombre incluyendo globales", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testCategoryName, testEmpresaID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testCategoryID))

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), testCategoryName, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testCategoryID, categoryID)
	})

	t.Run("devuelve ErrNotFound si no existe", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testCategoryName, testEmpresaID).
			WillReturnError(sql.ErrNoRows)

		categoryID, err := ms.store.GetCategoryIDByName(context.Background(), testCategoryName, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, categoryID)
	})
}
