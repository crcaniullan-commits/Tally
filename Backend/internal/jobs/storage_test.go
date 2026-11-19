package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBTest = errors.New("error de base de datos simulado")

type mockStore struct {
	store *AlmacenJobs
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

	ms.store = NuevoAlmacenJobs(db)
	ms.mock = mock

	return ms
}

func TestAlmacenJobs_UsuariosSinMovimientos(t *testing.T) {
	ctx := context.Background()
	start := time.Now().UTC()
	end := start.Add(24 * time.Hour)

	t.Run("obtiene lista de usuarios sin movimientos", func(t *testing.T) {
		ms := newMockStore(t)
		id1 := uuid.New()
		id2 := uuid.New()

		rows := sqlmock.NewRows([]string{"user_id"}).
			AddRow(id1).
			AddRow(id2)

		ms.mock.ExpectQuery("SELECT DISTINCT dt.user_id FROM device_tokens dt").
			WithArgs(start, end).
			WillReturnRows(rows)

		userIDs, err := ms.store.UsuariosSinMovimientos(ctx, start, end)
		assert.NoError(t, err)
		assert.Equal(t, []uuid.UUID{id1, id2}, userIDs)
		assert.Contains(t, ms.sql, "NOT EXISTS")
	})

	t.Run("propaga error de bd", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectQuery("SELECT DISTINCT dt.user_id FROM device_tokens dt").
			WithArgs(start, end).
			WillReturnError(errDBTest)

		userIDs, err := ms.store.UsuariosSinMovimientos(ctx, start, end)
		assert.ErrorIs(t, err, errDBTest)
		assert.Nil(t, userIDs)
	})
}

func TestAlmacenJobs_CobrosVencidosPorUsuario(t *testing.T) {
	ctx := context.Background()
	day := "2026-11-18"

	t.Run("obtiene resumen de cobros vencidos", func(t *testing.T) {
		ms := newMockStore(t)
		uid := uuid.New()

		rows := sqlmock.NewRows([]string{"user_id", "cantidad", "total"}).
			AddRow(uid, 3, int64(45000))

		ms.mock.ExpectQuery("SELECT e.user_id, COUNT").
			WithArgs(day).
			WillReturnRows(rows)

		res, err := ms.store.CobrosVencidosPorUsuario(ctx, day)
		assert.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, uid, res[0].UserID)
		assert.Equal(t, 3, res[0].Cantidad)
		assert.Equal(t, int64(45000), res[0].Total)
		assert.Contains(t, ms.sql, "GROUP BY e.user_id")
	})

	t.Run("propaga error de bd", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectQuery("SELECT e.user_id, COUNT").
			WithArgs(day).
			WillReturnError(errDBTest)

		res, err := ms.store.CobrosVencidosPorUsuario(ctx, day)
		assert.ErrorIs(t, err, errDBTest)
		assert.Nil(t, res)
	})
}
