package goals

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// mockStore envuelve un AlmacenMetas sobre sqlmock y guarda el SQL que el código
// realmente mandó a la base. sqlmock no parsea SQL, así que además auditamos el
// texto capturado a mano.
type mockStore struct {
	store *AlmacenMetas
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

	ms.store = NuevoAlmacenMetas(db)
	ms.mock = mock

	return ms
}

func goalColumns() []string {
	return []string{"id", "empresa_id", "nombre", "periodo", "monto_meta", "created_at"}
}

func goalRow() []driver.Value {
	return []driver.Value{
		testGoalID,
		testEmpresaID,
		"Ahorrar",
		string(model.GoalPeriodMes),
		int64(10000),
		testGoalCreatedAt,
	}
}

func TestStoreGoal_CreateGoal(t *testing.T) {
	t.Run("inserta el goal y devuelve los datos generados por la base", func(t *testing.T) {
		ms := newMockStore(t)
		goal := &model.Goal{
			EmpresaID: testEmpresaID,
			Nombre:    "Ahorrar",
			Periodo:   model.GoalPeriodMes,
			MontoMeta: 10000,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, goal.Nombre, goal.Periodo, goal.MontoMeta).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
				AddRow(testGoalID, testGoalCreatedAt))

		err := ms.store.CrearMeta(context.Background(), goal)

		require.NoError(t, err)
		assert.Equal(t, testGoalID, goal.ID)
		assert.Equal(t, testGoalCreatedAt, goal.CreatedAt)
		assert.Contains(t, ms.sql, "INSERT INTO goals")
		assert.Contains(t, ms.sql, "RETURNING id, created_at")
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)
		goal := &model.Goal{
			EmpresaID: testEmpresaID,
			Nombre:    "Ahorrar",
			Periodo:   model.GoalPeriodMes,
			MontoMeta: 10000,
		}

		ms.mock.ExpectQuery("cualquier INSERT").
			WithArgs(testEmpresaID, goal.Nombre, goal.Periodo, goal.MontoMeta).
			WillReturnError(errDBBoom)

		err := ms.store.CrearMeta(context.Background(), goal)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, uuid.Nil, goal.ID)
	})
}

func TestStoreGoal_DeleteGoal(t *testing.T) {
	t.Run("elimina el goal", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testGoalID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.EliminarMeta(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Contains(t, ms.sql, "DELETE FROM goals")
		assert.Contains(t, ms.sql, "WHERE id = $1 and empresa_id = $2")
	})

	t.Run("devuelve utilidades.ErrorNoEncontrado si no elimino ninguna fila", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testGoalID, testEmpresaID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.EliminarMeta(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testGoalID, testEmpresaID).
			WillReturnError(errDBBoom)

		err := ms.store.EliminarMeta(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("propaga el error al leer las filas afectadas", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectExec("cualquier DELETE").
			WithArgs(testGoalID, testEmpresaID).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := ms.store.EliminarMeta(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestStoreGoal_GetIncomesTotalByPeriod(t *testing.T) {
	casos := []struct {
		periodo model.GoalPeriod
		unit    string
	}{
		{model.GoalPeriodDia, "day"},
		{model.GoalPeriodSemana, "week"},
		{model.GoalPeriodMes, "month"},
	}

	for _, caso := range casos {
		t.Run(string(caso.periodo), func(t *testing.T) {
			ms := newMockStore(t)

			ms.mock.ExpectQuery("cualquier SELECT").
				WithArgs(testEmpresaID, caso.unit).
				WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(int64(2500)))

			total, err := ms.store.ObtenerTotalIngresosPorPeriodo(context.Background(), testEmpresaID, caso.periodo)

			require.NoError(t, err)
			assert.Equal(t, int64(2500), total)
			assert.Contains(t, ms.sql, "COALESCE(SUM(monto), 0)")
			assert.Contains(t, ms.sql, "FROM incomes")
			assert.Contains(t, ms.sql, "empresa_id = $1")
			assert.Contains(t, ms.sql, "date_trunc($2::text")
		})
	}

	t.Run("devuelve cero si no hay ingresos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, "day").
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(int64(0)))

		total, err := ms.store.ObtenerTotalIngresosPorPeriodo(context.Background(), testEmpresaID, model.GoalPeriodDia)

		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
	})

	t.Run("rechaza un periodo invalido sin consultar la base", func(t *testing.T) {
		ms := newMockStore(t)

		total, err := ms.store.ObtenerTotalIngresosPorPeriodo(context.Background(), testEmpresaID, model.GoalPeriod("anio"))

		require.Error(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, ms.sql)
	})

	t.Run("propaga el error de la base de datos", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, "week").
			WillReturnError(errDBBoom)

		total, err := ms.store.ObtenerTotalIngresosPorPeriodo(context.Background(), testEmpresaID, model.GoalPeriodSemana)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, int64(0), total)
	})
}

func TestStoreGoal_GetGoal(t *testing.T) {
	t.Run("devuelve el goal de la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testGoalID).
			WillReturnRows(sqlmock.NewRows(goalColumns()).AddRow(goalRow()...))

		goal, err := ms.store.ObtenerMeta(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testGoalID, goal.ID)
		assert.Equal(t, testEmpresaID, goal.EmpresaID)
		assert.Equal(t, "Ahorrar", goal.Nombre)
		assert.Equal(t, model.GoalPeriodMes, goal.Periodo)
		assert.Equal(t, int64(10000), goal.MontoMeta)
		assert.Equal(t, testGoalCreatedAt, goal.CreatedAt)
		assert.Contains(t, ms.sql, "WHERE empresa_id = $1 and id = $2")
	})

	t.Run("traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testGoalID).
			WillReturnError(sql.ErrNoRows)

		goal, err := ms.store.ObtenerMeta(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Equal(t, model.Goal{}, goal)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID, testGoalID).
			WillReturnError(errDBBoom)

		goal, err := ms.store.ObtenerMeta(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, model.Goal{}, goal)
	})
}

func TestStoreGoal_GetGoals(t *testing.T) {
	t.Run("devuelve todos los goals de la empresa", func(t *testing.T) {
		ms := newMockStore(t)

		otro := []driver.Value{
			uuid.MustParse("e5f6a7b8-9c0d-4e1f-8a2b-3c4d5e6f7a8b"),
			testEmpresaID,
			"Viaje",
			string(model.GoalPeriodSemana),
			int64(50000),
			testGoalCreatedAt,
		}

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(goalColumns()).
				AddRow(goalRow()...).
				AddRow(otro...))

		goals, err := ms.store.ObtenerMetas(context.Background(), testEmpresaID)

		require.NoError(t, err)
		require.Len(t, goals, 2)
		assert.Equal(t, testGoalID, goals[0].ID)
		assert.Equal(t, model.GoalPeriodSemana, goals[1].Periodo)
		assert.Contains(t, ms.sql, "WHERE empresa_id = $1")
	})

	t.Run("devuelve una lista vacia no nil si no hay goals", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(goalColumns()))

		goals, err := ms.store.ObtenerMetas(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.NotNil(t, goals)
		assert.Empty(t, goals)
	})

	t.Run("propaga el error de la query", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnError(errDBBoom)

		goals, err := ms.store.ObtenerMetas(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, goals)
	})

	t.Run("propaga el error durante la iteracion", func(t *testing.T) {
		ms := newMockStore(t)

		ms.mock.ExpectQuery("cualquier SELECT").
			WithArgs(testEmpresaID).
			WillReturnRows(sqlmock.NewRows(goalColumns()).
				AddRow(goalRow()...).
				RowError(0, errDBBoom))

		goals, err := ms.store.ObtenerMetas(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, goals)
	})
}
