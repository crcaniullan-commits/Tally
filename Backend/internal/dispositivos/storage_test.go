package dispositivos

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBTest = errors.New("error de base de datos simulado")

type mockStore struct {
	store *AlmacenDispositivos
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

	ms.store = NuevoAlmacenDispositivos(db)
	ms.mock = mock

	return ms
}

func TestAlmacenDispositivos_Upsert(t *testing.T) {
	ctx := context.Background()

	t.Run("upsert exitoso", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("INSERT INTO device_tokens").
			WithArgs(sqlmock.AnyArg(), testUserID, testToken, "android").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := ms.store.Upsert(ctx, testUserID, testToken, "android")
		assert.NoError(t, err)
		assert.Contains(t, ms.sql, "ON CONFLICT (token) DO UPDATE")
	})

	t.Run("propaga error de bd", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("INSERT INTO device_tokens").
			WithArgs(sqlmock.AnyArg(), testUserID, testToken, "android").
			WillReturnError(errDBTest)

		err := ms.store.Upsert(ctx, testUserID, testToken, "android")
		assert.ErrorIs(t, err, errDBTest)
	})
}

func TestAlmacenDispositivos_Eliminar(t *testing.T) {
	ctx := context.Background()

	t.Run("elimina dispositivo exitosamente", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM device_tokens").
			WithArgs(testToken, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := ms.store.Eliminar(ctx, testUserID, testToken)
		assert.NoError(t, err)
		assert.Contains(t, ms.sql, "token = $1 AND user_id = $2")
	})

	t.Run("retorna no encontrado si 0 filas afectadas", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM device_tokens").
			WithArgs(testToken, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := ms.store.Eliminar(ctx, testUserID, testToken)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})

	t.Run("propaga error de bd al eliminar", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM device_tokens").
			WithArgs(testToken, testUserID).
			WillReturnError(errDBTest)

		err := ms.store.Eliminar(ctx, testUserID, testToken)
		assert.ErrorIs(t, err, errDBTest)
	})
}

func TestAlmacenDispositivos_ObtenerTokensPorUsuario(t *testing.T) {
	ctx := context.Background()

	t.Run("obtiene lista de tokens", func(t *testing.T) {
		ms := newMockStore(t)
		rows := sqlmock.NewRows([]string{"token"}).
			AddRow(testToken).
			AddRow("otro-token")

		ms.mock.ExpectQuery("SELECT token FROM device_tokens").
			WithArgs(testUserID).
			WillReturnRows(rows)

		tokens, err := ms.store.ObtenerTokensPorUsuario(ctx, testUserID)
		assert.NoError(t, err)
		assert.Equal(t, []string{testToken, "otro-token"}, tokens)
	})

	t.Run("propaga error al consultar tokens", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectQuery("SELECT token FROM device_tokens").
			WithArgs(testUserID).
			WillReturnError(errDBTest)

		tokens, err := ms.store.ObtenerTokensPorUsuario(ctx, testUserID)
		assert.ErrorIs(t, err, errDBTest)
		assert.Nil(t, tokens)
	})
}

func TestAlmacenDispositivos_ListarPorUsuario(t *testing.T) {
	ctx := context.Background()

	t.Run("lista dispositivos del usuario", func(t *testing.T) {
		ms := newMockStore(t)
		now := time.Now().UTC()
		rows := sqlmock.NewRows([]string{"id", "user_id", "token", "platform", "last_seen_at"}).
			AddRow(testDeviceID, testUserID, testToken, "android", now)

		ms.mock.ExpectQuery("SELECT id, user_id, token, platform, last_seen_at FROM device_tokens").
			WithArgs(testUserID).
			WillReturnRows(rows)

		dispositivos, err := ms.store.ListarPorUsuario(ctx, testUserID)
		assert.NoError(t, err)
		require.Len(t, dispositivos, 1)
		assert.Equal(t, testDeviceID, dispositivos[0].ID)
		assert.Equal(t, testUserID, dispositivos[0].UserID)
		assert.Equal(t, testToken, dispositivos[0].Token)
		assert.Equal(t, "android", dispositivos[0].Platform)
		assert.Equal(t, now, dispositivos[0].LastSeenAt)
	})

	t.Run("propaga error al listar", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectQuery("SELECT id, user_id, token, platform, last_seen_at FROM device_tokens").
			WithArgs(testUserID).
			WillReturnError(errDBTest)

		dispositivos, err := ms.store.ListarPorUsuario(ctx, testUserID)
		assert.ErrorIs(t, err, errDBTest)
		assert.Nil(t, dispositivos)
	})
}

func TestAlmacenDispositivos_EliminarTokens(t *testing.T) {
	ctx := context.Background()

	t.Run("no hace nada si slice está vacío", func(t *testing.T) {
		ms := newMockStore(t)
		err := ms.store.EliminarTokens(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("elimina tokens por array", func(t *testing.T) {
		ms := newMockStore(t)
		ms.mock.ExpectExec("DELETE FROM device_tokens WHERE token = ANY").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(driver.RowsAffected(2))

		err := ms.store.EliminarTokens(ctx, []string{"t1", "t2"})
		assert.NoError(t, err)
		assert.Contains(t, ms.sql, "token = ANY($1)")
	})
}
