package incomes

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// mockStore envuelve un StoreIncome sobre sqlmock y guarda el SQL que el
// código realmente mandó a la base.
//
// sqlmock no parsea el SQL: solo compara strings y devuelve filas. Eso impide
// que detecte solo (p.ej.) un Scan desalineado con el SELECT. Por eso el
// matcher es permisivo y además auditamos el SQL capturado a mano.
type mockStore struct {
	store *StoreIncome
	mock  sqlmock.Sqlmock
	sql   string // última query realmente ejecutada
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

// columnas reales de la tabla incomes segun migrations/000001
func incomeColumns() []string {
	return []string{"id", "user_id", "monto", "payment_method", "descripcion", "fecha", "created_at"}
}

func incomeRow() []driver.Value {
	descripcion := "Venta de almuerzo"

	return []driver.Value{
		testIncomeID,
		testUserID,
		int64(1500),
		string(util.PaymentMethodDebito),
		descripcion,
		testFecha,
		testFecha,
	}
}

func TestStoreIncome_DeleteIncome(t *testing.T) {
	t.Run("elimina el ingreso filtrando por id y user_id", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testIncomeID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "DELETE FROM incomes")
		assert.Contains(t, ms.sql, "WHERE id = $1 and user_id = $2",
			"el delete debe acotar por usuario, si no es un IDOR")
	})

	t.Run("devuelve util.ErrNotFound si no se elimino ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testIncomeID, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base en vez de reportar un delete exitoso", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testIncomeID, testUserID).
			WillReturnError(errDBBoom)

		err := ms.store.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testIncomeID, testUserID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.DeleteIncome(context.Background(), testIncomeID, testUserID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("no puede borrar el ingreso de otro usuario", func(t *testing.T) {
		// El WHERE lleva user_id, asi que borrar un incomeID ajeno no matchea
		// ninguna fila y el store responde ErrNotFound en vez de 200.
		ms := newMockStore(t)

		otroUsuario := uuid.New()

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testIncomeID, otroUsuario).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.DeleteIncome(context.Background(), testIncomeID, otroUsuario)

		require.ErrorIs(t, err, util.ErrNotFound)
	})
}

func TestStoreIncome_GetAllIncomesOfUser(t *testing.T) {
	t.Run("mapea las 7 columnas a IncomeStorage", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(incomeColumns()).AddRow(incomeRow()...))

		incomes, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		require.Len(t, incomes, 1)

		assert.Equal(t, testIncomeID, incomes[0].ID)
		assert.Equal(t, testUserID, incomes[0].UserID)
		assert.Equal(t, int64(1500), incomes[0].Monto)
		assert.Equal(t, util.PaymentMethodDebito, incomes[0].PaymentMethod)
		require.NotNil(t, incomes[0].Descripcion)
		assert.Equal(t, "Venta de almuerzo", *incomes[0].Descripcion)
		assert.True(t, testFecha.Equal(incomes[0].Fecha))
	})

	t.Run("acumula varias filas", func(t *testing.T) {
		ms := newMockStore(t)

		segunda := incomeRow()
		otroID := uuid.New()
		segunda[0] = otroID
		segunda[2] = int64(99)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(incomeColumns()).
				AddRow(incomeRow()...).
				AddRow(segunda...))

		incomes, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		require.Len(t, incomes, 2)
		assert.Equal(t, testIncomeID, incomes[0].ID)
		assert.Equal(t, otroID, incomes[1].ID)
		assert.Equal(t, int64(99), incomes[1].Monto)
	})

	t.Run("acepta descripcion nula", func(t *testing.T) {
		ms := newMockStore(t)

		fila := incomeRow()
		fila[4] = nil

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(incomeColumns()).AddRow(fila...))

		incomes, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		require.Len(t, incomes, 1)
		assert.Nil(t, incomes[0].Descripcion)
	})

	t.Run("devuelve slice vacio y no nil cuando el usuario no tiene ingresos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(incomeColumns()))

		incomes, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)

		require.NoError(t, err)
		assert.NotNil(t, incomes, "un slice nil se serializa como null, no como []")
		assert.Empty(t, incomes)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		incomes, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, incomes)
	})

	t.Run("filtra por la columna user_id del esquema", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(incomeColumns()))

		_, err := ms.store.GetAllIncomesOfUser(context.Background(), testUserID)
		require.NoError(t, err)

		assert.Contains(t, ms.sql, "WHERE user_id = $1",
			"migrations/000001 creo la columna como user_id")
	})
}

func TestStoreIncome_AddIncome(t *testing.T) {
	t.Run("inserta en las columnas reales del esquema", func(t *testing.T) {
		ms := newMockStore(t)

		income := newTestIncome()
		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testUserID, income.Monto, income.PaymentMethod, income.Descripcion).
			WillReturnRows(sqlmock.NewRows([]string{"id", "fecha", "created_at"}).
				AddRow(testIncomeID, testFecha, testFecha))

		err := ms.store.AddIncome(context.Background(), &income)
		require.NoError(t, err)

		assert.Contains(t, ms.sql, "user_id, monto, payment_method, descripcion",
			"migrations/000001 creo user_id y payment_method")
		assert.NotContains(t, strings.ToLower(ms.sql), "patment")
	})

	t.Run("usa RETURNING para obtener el id y las fechas", func(t *testing.T) {
		ms := newMockStore(t)

		income := newTestIncome()
		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testUserID, income.Monto, income.PaymentMethod, income.Descripcion).
			WillReturnRows(sqlmock.NewRows([]string{"id", "fecha", "created_at"}))

		err := ms.store.AddIncome(context.Background(), &income)

		require.ErrorIs(t, err, sql.ErrNoRows,
			"sin filas devueltas el Scan debe fallar, no inventar un ID")
		assert.Contains(t, strings.ToUpper(ms.sql), "RETURNING",
			"sin RETURNING el alta nunca podria devolver el id generado")
	})

	t.Run("escribe en el income recibido el id y las fechas del RETURNING", func(t *testing.T) {
		// La firma recibe *IncomeStorage, asi que el Scan debe mutatear el
		// struct del service y no una copia. Si volviera a ser por valor, el
		// service devolveria un ingreso con uuid.Nil y fechas cero.
		ms := newMockStore(t)

		income := newTestIncome()
		income.ID = uuid.Nil
		income.Fecha = time.Time{}
		income.CreatedAt = time.Time{}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testUserID, income.Monto, income.PaymentMethod, income.Descripcion).
			WillReturnRows(sqlmock.NewRows([]string{"id", "fecha", "created_at"}).
				AddRow(testIncomeID, testFecha, testFecha))

		require.NoError(t, ms.store.AddIncome(context.Background(), &income))

		assert.Equal(t, testIncomeID, income.ID)
		assert.True(t, testFecha.Equal(income.Fecha))
		assert.True(t, testFecha.Equal(income.CreatedAt))
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		ms := newMockStore(t)

		income := newTestIncome()
		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testUserID, income.Monto, income.PaymentMethod, income.Descripcion).
			WillReturnError(errDBBoom)

		err := ms.store.AddIncome(context.Background(), &income)

		require.ErrorIs(t, err, errDBBoom)
	})
}
