package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

func newMockStore(t *testing.T) (*AlmacenAuth, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		mock.ExpectClose()
		_ = db.Close()
	})

	return NuevoAlmacenAuth(db), mock
}

var (
	regexInsertUsers = regexp.QuoteMeta("INSERT INTO users")
	regexSelectEmail = regexp.QuoteMeta("SELECT id, email, nombre, rut, password_hash")
	regexUpdateUsers = regexp.QuoteMeta("UPDATE users")
)

var testCreatedAt = time.Date(2027, time.January, 15, 12, 0, 0, 0, time.UTC)

func TestStoreAuth_Create(t *testing.T) {
	t.Run("inserta el usuario y devuelve id y created_at", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(user.Email, user.PasswordHash.GetHash(), user.Nombre, user.Role, user.Rut.String()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(testUserID, testCreatedAt))

		require.NoError(t, store.Create(context.Background(), user))
		assert.Equal(t, testUserID, user.ID)
		assert.True(t, testCreatedAt.Equal(user.CreatedAt))
	})

	t.Run("traduce el duplicado de email a users.ErrDuplicateEmail", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`))

		err := store.Create(context.Background(), newTestUser())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
	})

	t.Run("traduce el duplicado de rut a users.ErrDuplicateRut", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_rut_key"`))

		err := store.Create(context.Background(), newTestUser())

		require.ErrorIs(t, err, users.ErrDuplicateRut)
	})

	t.Run("propaga cualquier otro error de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errDBBoom)

		err := store.Create(context.Background(), newTestUser())

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("inserta por la transaccion del contexto si hay una abierta", func(t *testing.T) {
		// El INSERT tiene que salir por el *sql.Tx del contexto para que el
		// rollback de RegistrarUsuario pueda revertir la creacion del usuario.
		db, sqlm, err := sqlmock.New()
		require.NoError(t, err)

		t.Cleanup(func() {
			if err := sqlm.ExpectationsWereMet(); err != nil {
				t.Errorf("expectativas de sql sin cumplir: %v", err)
			}
			_ = db.Close()
		})

		user := newTestUser()

		sqlm.ExpectBegin()
		sqlm.ExpectQuery(regexInsertUsers).
			WithArgs(user.Email, user.PasswordHash.GetHash(), user.Nombre, user.Role, user.Rut.String()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(testUserID, testCreatedAt))
		sqlm.ExpectCommit()

		tx, err := db.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		require.NoError(t, NuevoAlmacenAuth(db).Create(bdtx.WithTx(context.Background(), tx), user))
		require.NoError(t, tx.Commit())
	})
}

func TestStoreAuth_GetByEmail(t *testing.T) {
	emailColumns := func() []string {
		return []string{"id", "email", "nombre", "rut", "password_hash"}
	}

	t.Run("devuelve el usuario con el rut parseado", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectQuery(regexSelectEmail).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows(emailColumns()).
				AddRow(testUserID, "emprendedor@correo.cl", "Nombre Original", "19.234.567-K", user.PasswordHash.GetHash()))

		got, err := store.ObtenerPorEmail(context.Background(), "emprendedor@correo.cl")

		require.NoError(t, err)
		assert.Equal(t, testUserID, got.ID)
		assert.Equal(t, "Nombre Original", got.Nombre)
		assert.Equal(t, utilidades.RUT{Cuerpo: 19234567, DV: "K"}, got.Rut)
	})

	t.Run("traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexSelectEmail).
			WithArgs("nadie@correo.cl").
			WillReturnRows(sqlmock.NewRows(emailColumns()))

		got, err := store.ObtenerPorEmail(context.Background(), "nadie@correo.cl")

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Nil(t, got)
	})

	t.Run("propaga el error si el rut guardado no tiene formato", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexSelectEmail).
			WithArgs("emprendedor@correo.cl").
			WillReturnRows(sqlmock.NewRows(emailColumns()).
				AddRow(testUserID, "emprendedor@correo.cl", "Nombre Original", "rut-corrupto", []byte("hash")))

		_, err := store.ObtenerPorEmail(context.Background(), "emprendedor@correo.cl")

		require.ErrorIs(t, err, utilidades.ErrFormatoInvalido)
	})
}

// establecerVencimiento escribe plan_expires_at (migrations/000004). No devuelve
// utilidades.ErrorNoEncontrado si no actualiza ninguna fila: la transaccion que envuelve la
// llamada ya revierte ante cualquier error.
func TestStoreAuth_setExpire(t *testing.T) {
	t.Run("fija plan_expires_at del usuario", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdateUsers).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.establecerVencimiento(context.Background(), expire, testUserID))
	})

	t.Run("envuelve el error de la base con el nombre de la operacion", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdateUsers).
			WithArgs(expire, testUserID).
			WillReturnError(errDBBoom)

		err := store.establecerVencimiento(context.Background(), expire, testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Contains(t, err.Error(), "plan_expires_at")
	})

	t.Run("escribe por la transaccion del contexto si hay una abierta", func(t *testing.T) {
		db, sqlm, err := sqlmock.New()
		require.NoError(t, err)

		t.Cleanup(func() {
			if err := sqlm.ExpectationsWereMet(); err != nil {
				t.Errorf("expectativas de sql sin cumplir: %v", err)
			}
			_ = db.Close()
		})

		expire := time.Now().Add(24 * time.Hour)

		sqlm.ExpectBegin()
		sqlm.ExpectExec(regexUpdateUsers).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		sqlm.ExpectCommit()

		tx, err := db.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		require.NoError(t, NuevoAlmacenAuth(db).establecerVencimiento(bdtx.WithTx(context.Background(), tx), expire, testUserID))
		require.NoError(t, tx.Commit())
	})
}

func TestStoreAuth_setExpire_query(t *testing.T) {
	newCapturingStore := func(t *testing.T) (*AlmacenAuth, sqlmock.Sqlmock, *string) {
		t.Helper()

		var capturada string

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
			sqlmock.QueryMatcherFunc(func(_, actualSQL string) error {
				capturada = actualSQL
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

		return NuevoAlmacenAuth(db), mock, &capturada
	}

	t.Run("escribe en users.plan_expires_at, no en expires_at", func(t *testing.T) {
		store, mock, sql := newCapturingStore(t)

		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))

		expire := time.Now().Add(24 * time.Hour)
		require.NoError(t, store.establecerVencimiento(context.Background(), expire, testUserID))

		assert.Contains(t, *sql, "SET plan_expires_at = $1")
		assert.NotRegexp(t, `\bexpires_at\b`, *sql, "la query no debe tocar expires_at")
	})

	t.Run("filtra por id, no por el codigo de una llave", func(t *testing.T) {
		store, mock, sql := newCapturingStore(t)

		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))

		expire := time.Now().Add(24 * time.Hour)
		require.NoError(t, store.establecerVencimiento(context.Background(), expire, testUserID))

		assert.Contains(t, *sql, "WHERE id = $2")
		assert.NotContains(t, *sql, "code")
	})
}
