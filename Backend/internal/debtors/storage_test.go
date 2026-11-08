package debtors

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// mockStore envuelve un StoreDebtor sobre sqlmock y guarda el SQL que el
// código realmente mandó a la base. sqlmock no parsea SQL, así que además
// auditamos el texto capturado a mano.
type mockStore struct {
	store *StoreDebtor
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

func debtorColumns() []string {
	return []string{
		"id",
		"empresa_id",
		"nombre",
		"monto",
		"descripcion",
		"pagado",
		"fecha_limite",
		"created_at",
	}
}

func debtorRow() []driver.Value {
	return []driver.Value{
		testDebtorID,
		testEmpresaID,
		"Cliente",
		int64(15000),
		"Deuda pendiente",
		false,
		testFechaLimite,
		testDebtorCreatedAt,
	}
}

func TestStoreDebtor_GetDebtor(t *testing.T) {
	t.Run("devuelve si la deuda esta pagada", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnRows(sqlmock.NewRows([]string{"pagado"}).AddRow(true))

		pagado, err := ms.store.GetDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.NoError(t, err)
		assert.True(t, pagado)
		assert.Contains(t, ms.sql, "SELECT pagado FROM debtors")
		assert.Contains(t, ms.sql, "WHERE id = $1 and empresa_id = $2")
	})

	t.Run("devuelve false si la deuda no esta pagada", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnRows(sqlmock.NewRows([]string{"pagado"}).AddRow(false))

		pagado, err := ms.store.GetDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.NoError(t, err)
		assert.False(t, pagado)
	})

	t.Run("traduce sql.ErrNoRows a util.ErrNotFound", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnError(sql.ErrNoRows)

		pagado, err := ms.store.GetDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.False(t, pagado)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnError(errDBBoom)

		pagado, err := ms.store.GetDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.False(t, pagado)
	})
}

func TestStoreDebtor_Create(t *testing.T) {
	t.Run("inserta la deuda y devuelve los datos generados por la base", func(t *testing.T) {
		ms := newMockStore(t)
		debtor := newTestDebtor()
		debtor.ID = uuid.Nil
		debtor.CreatedAt = testDebtorCreatedAt

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(
				debtor.EmpresaID,
				debtor.Nombre,
				debtor.Monto,
				*debtor.Descripcion,
				*debtor.FechaLimite,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pagado", "created_at"}).
				AddRow(testDebtorID, false, testDebtorCreatedAt))

		err := ms.store.Create(context.Background(), &debtor)

		require.NoError(t, err)
		assert.Equal(t, testDebtorID, debtor.ID)
		assert.False(t, debtor.Pagado)
		assert.Equal(t, testDebtorCreatedAt, debtor.CreatedAt)
		assert.Contains(t, ms.sql, "INSERT INTO debtors")
		assert.Contains(t, ms.sql, "RETURNING id, pagado, created_at")
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)
		debtor := newTestDebtor()
		debtor.ID = uuid.Nil

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(
				debtor.EmpresaID,
				debtor.Nombre,
				debtor.Monto,
				*debtor.Descripcion,
				*debtor.FechaLimite,
			).
			WillReturnError(errDBBoom)

		err := ms.store.Create(context.Background(), &debtor)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, uuid.Nil, debtor.ID)
	})
}

func TestStoreDebtor_Update(t *testing.T) {
	debtor := newTestDebtor()

	t.Run("actualiza la deuda", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(
				debtor.Nombre,
				debtor.Monto,
				*debtor.FechaLimite,
				testDebtorID,
				testEmpresaID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.Update(context.Background(), testDebtorID, testEmpresaID, debtor)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "UPDATE debtors")
		assert.Contains(t, ms.sql, "SET nombre = $1, monto = $2, fecha_limite = $3")
		assert.Contains(t, ms.sql, "WHERE id = $4 and empresa_id = $5")
	})

	t.Run("devuelve util.ErrNotFound si no actualizo ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(
				debtor.Nombre,
				debtor.Monto,
				*debtor.FechaLimite,
				testDebtorID,
				testEmpresaID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.Update(context.Background(), testDebtorID, testEmpresaID, debtor)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(
				debtor.Nombre,
				debtor.Monto,
				*debtor.FechaLimite,
				testDebtorID,
				testEmpresaID,
			).
			WillReturnError(errDBBoom)

		err := ms.store.Update(context.Background(), testDebtorID, testEmpresaID, debtor)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error al leer las filas afectadas", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(
				debtor.Nombre,
				debtor.Monto,
				*debtor.FechaLimite,
				testDebtorID,
				testEmpresaID,
			).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.Update(context.Background(), testDebtorID, testEmpresaID, debtor)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreDebtor_PaidDebtor(t *testing.T) {
	t.Run("marca la deuda como pagada", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.PaidDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "UPDATE debtors")
		assert.Contains(t, ms.sql, "SET pagado = true")
		assert.Contains(t, ms.sql, "WHERE id = $1 and empresa_id = $2")
	})

	t.Run("devuelve util.ErrNotFound si no actualizo ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.PaidDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnError(errDBBoom)

		err := ms.store.PaidDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error al leer las filas afectadas", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier UPDATE").
			WithArgs(testDebtorID, testEmpresaID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.PaidDebtor(context.Background(), testDebtorID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreDebtor_GetDebtors(t *testing.T) {
	t.Run("devuelve todas las deudas de la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		segunda := []driver.Value{
			testOtherDebtorID,
			testEmpresaID,
			"Cliente 2",
			int64(20000),
			nil,
			true,
			nil,
			testDebtorCreatedAt,
		}

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(debtorColumns()).
				AddRow(debtorRow()...).
				AddRow(segunda...))

		debtors, err := ms.store.GetDebtors(context.Background(), testEmpresaID)

		require.NoError(t, err)
		require.Len(t, debtors, 2)

		assert.Equal(t, testDebtorID, debtors[0].ID)
		require.NotNil(t, debtors[0].Descripcion)
		assert.Equal(t, "Deuda pendiente", *debtors[0].Descripcion)
		require.NotNil(t, debtors[0].FechaLimite)
		assert.True(t, debtors[0].FechaLimite.Equal(testFechaLimite))

		assert.Equal(t, testOtherDebtorID, debtors[1].ID)
		assert.Nil(t, debtors[1].Descripcion)
		assert.Nil(t, debtors[1].FechaLimite)
		assert.True(t, debtors[1].Pagado)

		assert.Contains(t, ms.sql, "FROM debtors")
		assert.Contains(t, ms.sql, "WHERE empresa_id = $1")
		assert.Contains(t, ms.sql, "ORDER BY created_at DESC, id DESC")
	})

	t.Run("devuelve una lista vacia no nil si no hay deudas", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(debtorColumns()))

		debtors, err := ms.store.GetDebtors(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.NotNil(t, debtors)
		assert.Empty(t, debtors)
	})

	t.Run("propaga el error de la query", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnError(errDBBoom)

		debtors, err := ms.store.GetDebtors(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, debtors)
	})

	t.Run("propaga el error durante la iteracion", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(debtorColumns()).
				AddRow(debtorRow()...).
				RowError(0, errDBBoom))

		debtors, err := ms.store.GetDebtors(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, debtors)
	})
}
