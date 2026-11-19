package notificaciones

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBTest = errors.New("error de base de datos simulado")

type mockStore struct {
	store *AlmacenNotificaciones
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

	ms.store = NuevoAlmacenNotificaciones(db)
	ms.mock = mock

	return ms
}

func TestAlmacenNotificaciones_Claim(t *testing.T) {
	ctx := context.Background()

	t.Run("obtiene candado exitosamente (1 fila insertada)", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("INSERT INTO notification_log").
			WithArgs(testUserID, "daily", "2026-11-18").
			WillReturnResult(driver.RowsAffected(1))

		ok, err := ms.store.Claim(ctx, testUserID, "daily", "2026-11-18")
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Contains(t, ms.sql, "ON CONFLICT DO NOTHING")
	})

	t.Run("conflicto: candado ya tomado (0 filas insertadas)", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("INSERT INTO notification_log").
			WithArgs(testUserID, "daily", "2026-11-18").
			WillReturnResult(driver.RowsAffected(0))

		ok, err := ms.store.Claim(ctx, testUserID, "daily", "2026-11-18")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("error de base de datos en claim", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("INSERT INTO notification_log").
			WithArgs(testUserID, "daily", "2026-11-18").
			WillReturnError(errDBTest)

		ok, err := ms.store.Claim(ctx, testUserID, "daily", "2026-11-18")
		assert.ErrorIs(t, err, errDBTest)
		assert.False(t, ok)
	})
}

func TestAlmacenNotificaciones_Unclaim(t *testing.T) {
	ctx := context.Background()

	t.Run("elimina candado exitosamente", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM notification_log").
			WithArgs(testUserID, "daily", "2026-11-18").
			WillReturnResult(driver.RowsAffected(1))

		err := ms.store.Unclaim(ctx, testUserID, "daily", "2026-11-18")
		assert.NoError(t, err)
		assert.Contains(t, ms.sql, "user_id = $1 AND kind = $2 AND ref_date = $3")
	})

	t.Run("error de bd en unclaim", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM notification_log").
			WithArgs(testUserID, "daily", "2026-11-18").
			WillReturnError(errDBTest)

		err := ms.store.Unclaim(ctx, testUserID, "daily", "2026-11-18")
		assert.ErrorIs(t, err, errDBTest)
	})
}

func TestAlmacenNotificaciones_ListarPorUsuario(t *testing.T) {
	ctx := context.Background()

	t.Run("lista historial de notificaciones", func(t *testing.T) {
		ms := newMockStore(t)
		refDate := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
		sentAt := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"user_id", "kind", "ref_date", "sent_at"}).
			AddRow(testUserID, "daily", refDate, sentAt)

		ms.mock.ExpectQuery("SELECT user_id, kind, ref_date, sent_at FROM notification_log").
			WithArgs(testUserID).
			WillReturnRows(rows)

		logs, err := ms.store.ListarPorUsuario(ctx, testUserID)
		assert.NoError(t, err)
		require.Len(t, logs, 1)
		assert.Equal(t, testUserID, logs[0].UserID)
		assert.Equal(t, "daily", logs[0].Kind)
		assert.Equal(t, refDate, logs[0].RefDate)
		assert.Equal(t, sentAt, logs[0].SentAt)
	})

	t.Run("error al listar historial", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectQuery("SELECT user_id, kind, ref_date, sent_at FROM notification_log").
			WithArgs(testUserID).
			WillReturnError(errDBTest)

		logs, err := ms.store.ListarPorUsuario(ctx, testUserID)
		assert.ErrorIs(t, err, errDBTest)
		assert.Nil(t, logs)
	})
}
