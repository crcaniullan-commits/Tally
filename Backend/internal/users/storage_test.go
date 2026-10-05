package users

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// newMockStore levanta un UserStore sobre una conexión sqlmockeada.
func newMockStore(t *testing.T) (*UserStore, sqlmock.Sqlmock) {
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

	return NewStorage(db), mock
}

// Fragments de las queries reales, para no duplicar el SQL en cada test.
var (
	regexGetByID  = regexp.QuoteMeta("SELECT id, email, nombre, rut, role")
	regexGetByRut = regexp.QuoteMeta("SELECT id, email, nombre, rut, role")
	regexUpdate   = regexp.QuoteMeta("UPDATE users")
	regexDelete   = regexp.QuoteMeta("DELETE FROM users")
)

// columnas de la query de lectura: id, email, nombre, rut, role
func userColumns() []string {
	return []string{"id", "email", "nombre", "rut", "role"}
}

func userRow() []driver.Value {
	return []driver.Value{
		testUserID,
		"emprendedor@correo.cl",
		"Nombre Original",
		"19.234.567-K",
		string(util.UserRoleUsuario),
	}
}

func TestUserStore_GetByID(t *testing.T) {
	t.Run("devuelve el usuario con el rut parseado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(userRow()...))

		user, err := store.GetByID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, "emprendedor@correo.cl", user.Email)
		assert.Equal(t, "Nombre Original", user.Nombre)
		assert.Equal(t, util.UserRoleUsuario, user.Role)
		assert.Equal(t, util.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
	})

	t.Run("acepta el rut sin separadores", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := userRow()
		row[3] = "19234567K"

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(row...))

		user, err := store.GetByID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, util.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
	})

	t.Run("traduce sql.ErrNoRows a util.ErrNotFound", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumns()))

		user, err := store.GetByID(context.Background(), testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("devuelve util.ErrFormatoInvalido si el rut de la base no tiene formato", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := userRow()
		row[3] = "rut-corrupto"

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(row...))

		user, err := store.GetByID(context.Background(), testUserID)

		require.ErrorIs(t, err, util.ErrFormatoInvalido)
		assert.Nil(t, user)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		user, err := store.GetByID(context.Background(), testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, user)
	})
}

func TestUserStore_GetByRut(t *testing.T) {
	t.Run("busca por el rut recibido como string", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(userRow()...))

		user, err := store.GetByRut(context.Background(), "19.234.567-K")

		require.NoError(t, err)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, util.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
	})

	t.Run("traduce sql.ErrNoRows a util.ErrNotFound", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnRows(sqlmock.NewRows(userColumns()))

		user, err := store.GetByRut(context.Background(), "19.234.567-K")

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnError(errDBBoom)

		user, err := store.GetByRut(context.Background(), "19.234.567-K")

		require.ErrorIs(t, err, errDBBoom)
		assert.Nil(t, user)
	})
}

func TestUserStore_Update(t *testing.T) {
	t.Run("persiste hash y nombre del usuario", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()
		require.NoError(t, user.PasswordHash.Set("NuevaClave123"))

		mock.ExpectExec(regexUpdate).
			WithArgs(user.PasswordHash.GetHash(), user.Nombre, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(context.Background(), user)

		require.NoError(t, err)
	})

	t.Run("devuelve util.ErrNotFound si no se actualizo ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectExec(regexUpdate).
			WithArgs(user.PasswordHash.GetHash(), user.Nombre, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(context.Background(), user)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectExec(regexUpdate).
			WithArgs(user.PasswordHash.GetHash(), user.Nombre, user.ID).
			WillReturnError(errDBBoom)

		err := store.Update(context.Background(), user)

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, util.ErrNotFound, "un error de DB no es 'no encontrado'")
	})
}

func TestUserStore_Delete(t *testing.T) {
	t.Run("elimina el usuario con el id recibido", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(context.Background(), testUserID)

		require.NoError(t, err)
	})

	t.Run("devuelve util.ErrNotFound si no se elimino ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(context.Background(), testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		err := store.Delete(context.Background(), testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, util.ErrNotFound)
	})
}

func TestUserStore_ContextCancelado(t *testing.T) {
	t.Run("no ejecuta nada si el contexto ya esta cancelado", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := NewStorage(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = store.GetByID(ctx, testUserID)
		require.ErrorIs(t, err, context.Canceled)

		require.ErrorIs(t, store.Delete(ctx, testUserID), context.Canceled)
	})
}

// setExpire escribe plan_expires_at (migrations/000004), que es la unica
// columna de users que el service de canje toca. No devuelve util.ErrNotFound
// si no actualiza ninguna fila: el canje ya mostro que el usuario existe y la
// transaction que envuelve la llamada ya revierte ante cualquier error.
func TestUserStore_setExpire(t *testing.T) {
	t.Run("fija plan_expires_at del usuario", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.setExpire(context.Background(), expire, testUserID)

		require.NoError(t, err)
	})

	t.Run("no falla si el UPDATE no toca ninguna fila", func(t *testing.T) {
		// El store ya no mira RowsAffected (antes devolvia util.ErrNotFound).
		// Con un usuario inexistente el canje queda reportado como exitoso y el
		// commit hace su trabajo sin extender ningun plan.
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.setExpire(context.Background(), expire, testUserID)

		require.NoError(t, err)
	})

	t.Run("envuelve el error de la base con el nombre de la operacion", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnError(errDBBoom)

		err := store.setExpire(context.Background(), expire, testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Contains(t, err.Error(), "plan_expires_at")
		assert.NotErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("escribe por la transaccion del contexto si hay una abierta", func(t *testing.T) {
		// dbtx.FromContext hace que el canje y la extension del plan sean
		// atomicos: dentro de ExchangeCode la llamada tiene que salir por el
		// *sql.Tx del contexto, no por la conexion suelta.
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
		sqlm.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		sqlm.ExpectCommit()

		tx, err := db.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		require.NoError(t, NewStorage(db).setExpire(dbtx.WithTx(context.Background(), tx), expire, testUserID))
		require.NoError(t, tx.Commit())
	})
}

// newCapturingMockStore levanta el store con un matcher permisivo que guarda el
// SQL realmente ejecutado. sqlmock no parsea SQL (solo compara strings), asi que
// sin esto una coma de menos o una columna cambiada de lugar pasarian los tests
// en verde y reventarian recien contra Postgres.
func newCapturingMockStore(t *testing.T) (*UserStore, sqlmock.Sqlmock, *string) {
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

	return NewStorage(db), mock, &capturada
}

// columnasDe extrae la lista de proyeccion de un SELECT y la parte por coma.
// Sirve para comparar la query contra el orden en que el Scan escanea.
func columnasDe(query string) []string {
	query = strings.TrimSpace(query)
	query = query[strings.Index(query, "SELECT")+len("SELECT"):]
	query = query[:strings.Index(query, "FROM")]

	proyeccion := strings.Split(query, ",")
	for i, col := range proyeccion {
		proyeccion[i] = strings.TrimSpace(col)
	}

	return proyeccion
}

func TestUserStore_setExpire_query(t *testing.T) {
	t.Run("escribe en users.plan_expires_at, no en expires_at", func(t *testing.T) {
		// plan_expires_at es de users (migrations/000004) y expires_at es de
		// access_keys (migrations/000001): apuntar a la columna equivocada deja
		// al usuario sin plan.
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectExec("").
			WillReturnResult(sqlmock.NewResult(0, 1))

		expire := time.Now().Add(24 * time.Hour)
		require.NoError(t, store.setExpire(context.Background(), expire, testUserID))

		assert.Contains(t, *sql, "SET plan_expires_at = $1")
		assert.NotRegexp(t, `\bexpires_at\b`, *sql, "la query no debe tocar expires_at")
	})

	t.Run("filtra por id, no por el codigo de una llave", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectExec("").
			WillReturnResult(sqlmock.NewResult(0, 1))

		expire := time.Now().Add(24 * time.Hour)
		require.NoError(t, store.setExpire(context.Background(), expire, testUserID))

		assert.Contains(t, *sql, "WHERE id = $2")
		assert.NotContains(t, *sql, "code")
	})
}

// columnas de lectura que el Scan de GetByID / GetByRut realmente consume.
func userColumnsConPlan() []string {
	return append(userColumns(), "plan_expires_at")
}

// fila de lectura con las 6 columnas que devuelve el SELECT real.
func userRowConPlan() []driver.Value {
	return append(userRow(), time.Now().Add(24*time.Hour))
}

func TestUserStore_lectura_query(t *testing.T) {
	// El SELECT de GetByID y GetByRut pide plan_expires_at (migrations/000004)
	// pero el Scan solo lee 5 destinos. Contra Postgres eso es
	// "sql: expected 6 destination arguments in Scan, not 5" en TODAS las filas,
	// y como GetByID es lo que usa el middleware de autenticacion para resolver
	// el usuario del token, ningun endpoint bajo /v1/app llega a ejecutarse.
	// Los tests de arriba pasan porque(sqlmock) devuelven 5 columnas.
	t.Run("BUG: GetByID lee una columna de menos y falla siempre", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(userColumnsConPlan()).AddRow(userRowConPlan()...))

		user, err := store.GetByID(context.Background(), testUserID)

		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "expected 6 destination arguments in Scan, not 5")
		assert.Len(t, columnasDe(*sql), 6, "la proyeccion real pide 6 columnas")
	})

	t.Run("BUG: GetByRut lee una columna de menos y falla siempre", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(userColumnsConPlan()).AddRow(userRowConPlan()...))

		user, err := store.GetByRut(context.Background(), "19.234.567-K")

		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "expected 6 destination arguments in Scan, not 5")
		assert.Len(t, columnasDe(*sql), 6, "la proyeccion real pide 6 columnas")
	})
}
