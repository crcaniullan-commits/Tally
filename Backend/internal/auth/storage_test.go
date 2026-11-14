package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/lib/pq"
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
	regexInsertUsers      = regexp.QuoteMeta("INSERT INTO users")
	regexSelectEmail      = regexp.QuoteMeta("SELECT id, email, nombre, rut, password_hash")
	regexUpdateUsers      = regexp.QuoteMeta("UPDATE users")
	regexInsertCodigo     = regexp.QuoteMeta("INSERT INTO password_reset_codes")
	regexSelectCodigo     = regexp.QuoteMeta("SELECT id, user_id, codigo, expira_en, verificado_en, usado_en, created_at")
	regexInvalidarCodigos = regexp.QuoteMeta("UPDATE password_reset_codes")
	regexMarcarVerificado = regexp.QuoteMeta("SET verificado_en = now()")
	regexMarcarUsado      = regexp.QuoteMeta("SET usado_en = now()")
	regexUpdatePassword   = regexp.QuoteMeta("SET password_hash = $1")
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

	t.Run("traduce el duplicado de email aunque pq agregue el codigo SQL", func(t *testing.T) {
		// lib/pq devuelve el mensaje con el sufijo " (23505)": comparar el texto
		// exacto hacía que el error no matcheara y el handler respondía 500.
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_email_key" (23505)`))

		err := store.Create(context.Background(), newTestUser())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
	})

	t.Run("traduce el duplicado de email con el error tipado de pq", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(&pq.Error{Code: "23505", Constraint: "users_email_key"})

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

	t.Run("traduce el duplicado de rut aunque pq agregue el codigo SQL", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexInsertUsers).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(&pq.Error{Code: "23505", Constraint: "users_rut_key"})

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

func TestStoreAuth_CodigoRecuperacion(t *testing.T) {
	t.Run("invalida los codigos previos y guarda el nuevo", func(t *testing.T) {
		store, mock := newMockStore(t)

		codigo := &model.CodigoRecuperacion{
			UserID:   testUserID,
			Codigo:   hashearCodigo(testCodigoRecuperacion),
			ExpiraEn: time.Now().Add(15 * time.Minute),
		}

		mock.ExpectExec(regexInvalidarCodigos).
			WithArgs(testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery(regexInsertCodigo).
			WithArgs(testUserID, codigo.Codigo, codigo.ExpiraEn).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(testCodigoRecuperacionID, testCreatedAt))

		require.NoError(t, store.CrearCodigoRecuperacion(context.Background(), codigo))
		assert.Equal(t, testCodigoRecuperacionID, codigo.ID)
	})

	t.Run("propaga el error si no se puede invalidar los codigos previos", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexInvalidarCodigos).
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		err := store.CrearCodigoRecuperacion(context.Background(), &model.CodigoRecuperacion{UserID: testUserID})

		require.ErrorIs(t, err, errDBBoom)
	})

	t.Run("devuelve el ultimo codigo emitido", func(t *testing.T) {
		store, mock := newMockStore(t)

		expira := time.Now().Add(15 * time.Minute)

		mock.ExpectQuery(regexSelectCodigo).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "codigo", "expira_en", "verificado_en", "usado_en", "created_at"}).
				AddRow(testCodigoRecuperacionID, testUserID, hashearCodigo(testCodigoRecuperacion), expira, nil, nil, testCreatedAt))

		codigo, err := store.ObtenerCodigoRecuperacion(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, testCodigoRecuperacionID, codigo.ID)
		assert.Equal(t, hashearCodigo(testCodigoRecuperacion), codigo.Codigo)
		assert.Nil(t, codigo.VerificadoEn)
		assert.Nil(t, codigo.UsadoEn)
	})

	t.Run("traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexSelectCodigo).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "codigo", "expira_en", "verificado_en", "usado_en", "created_at"}))

		codigo, err := store.ObtenerCodigoRecuperacion(context.Background(), testUserID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Nil(t, codigo)
	})

	t.Run("marca el codigo como verificado y como usado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexMarcarVerificado).
			WithArgs(testCodigoRecuperacionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(regexMarcarUsado).
			WithArgs(testCodigoRecuperacionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.MarcarCodigoVerificado(context.Background(), testCodigoRecuperacionID))
		require.NoError(t, store.MarcarCodigoUsado(context.Background(), testCodigoRecuperacionID))
	})

	t.Run("actualiza el password del usuario", func(t *testing.T) {
		store, mock := newMockStore(t)

		hash := []byte("hash-de-prueba")

		mock.ExpectExec(regexUpdatePassword).
			WithArgs(hash, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.ActualizarPassword(context.Background(), testUserID, hash))
	})

	t.Run("devuelve ErrorNoEncontrado si el usuario no existe", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexUpdatePassword).
			WithArgs(sqlmock.AnyArg(), testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.ActualizarPassword(context.Background(), testUserID, []byte("hash"))

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})
}
