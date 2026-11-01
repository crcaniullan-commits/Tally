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
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, int64(1500), testCategoryID, &descripcion).
			WillReturnRows(sqlmock.NewRows([]string{"id", "fecha", "created_at"}).AddRow(testExpenseID, testFecha, testFecha.Add(2*time.Hour)))

		err := ms.store.AddExpense(context.Background(), expense)

		require.NoError(t, err)
		assert.Equal(t, testExpenseID, expense.ID)
		assert.Equal(t, testFecha, expense.Fecha)
		assert.Contains(t, ms.sql, "INSERT INTO expenses (empresa_id, monto, category_id, descripcion)",
			"la base asigna fecha con DEFAULT CURRENT_DATE")
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		expense := &model.Expense{
			EmpresaID:  testEmpresaID,
			Monto:      1500,
			CategoryID: testCategoryID,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, int64(1500), testCategoryID, (*string)(nil)).
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

func TestStoreExpense_CategoryExists(t *testing.T) {
	t.Run("acepta una categoria global o de la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testCategoryID, testEmpresaID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		existe, err := ms.store.CategoryExists(context.Background(), testCategoryID, testEmpresaID)

		require.NoError(t, err)
		assert.True(t, existe)
		assert.Contains(t, ms.sql, "FROM categories")
		assert.Contains(t, ms.sql, "id = $1 AND (empresa_id IS NULL OR empresa_id = $2)",
			"tiene que aceptar las globales y las propias de la empresa")
	})

	t.Run("devuelve false si la categoria no existe o es de otra empresa", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testCategoryID, testEmpresaID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		existe, err := ms.store.CategoryExists(context.Background(), testCategoryID, testEmpresaID)

		require.NoError(t, err)
		assert.False(t, existe)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testCategoryID, testEmpresaID).
			WillReturnError(errDBBoom)

		existe, err := ms.store.CategoryExists(context.Background(), testCategoryID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.False(t, existe)
	})
}
