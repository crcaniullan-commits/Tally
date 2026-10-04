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
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
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

// columnas de la query de GetExpire: id, code, created_by, redeemed_by,
// redeemed_at, expires_at, revoked_at, revoked_by, created_at
func accessKeyColumns() []string {
	return []string{
		"id", "code", "created_by", "redeemed_by",
		"redeemed_at", "expires_at", "revoked_at", "revoked_by", "created_at",
	}
}

// accessKeyRow devuelve una llave vigente y no canjeada. redeemed_by,
// redeemed_at, revoked_at y revoked_by van en NULL, que es el estado normal de
// una llave recien emitida.
func accessKeyRow(expiresAt time.Time) []driver.Value {
	return []driver.Value{
		testAccessKeyID,
		testAccessKeyCod,
		testMunicipalID,
		nil,
		nil,
		expiresAt,
		nil,
		nil,
		testAccessKeyCreatedAt,
	}
}

var (
	regexGetExpire         = regexp.QuoteMeta("SELECT id, code, created_by, redeemed_by")
	testAccessKeyCreatedAt = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
)

func TestUserStore_GetExpire(t *testing.T) {
	t.Run("devuelve la llave con las 9 columnas de access_keys", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectQuery(regexGetExpire).
			WithArgs(testAccessKeyCod).
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(accessKeyRow(expire)...))

		accessKey, err := store.GetExpire(context.Background(), testAccessKeyCod)

		require.NoError(t, err)
		assert.Equal(t, testAccessKeyID, accessKey.ID)
		assert.Equal(t, testAccessKeyCod, accessKey.Code)
		assert.Equal(t, testMunicipalID, accessKey.CreatedBy)
		assert.Nil(t, accessKey.RedeemedBy)
		assert.Nil(t, accessKey.RedeemedAt)
		require.NotNil(t, accessKey.ExpiresAt)
		assert.True(t, expire.Equal(*accessKey.ExpiresAt))
		assert.Nil(t, accessKey.RevokedAt)
		assert.Nil(t, accessKey.RevokedBy)
		assert.True(t, testAccessKeyCreatedAt.Equal(accessKey.CreatedAt))
	})

	t.Run("traduce sql.ErrNoRows a util.ErrNotFound", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetExpire).
			WithArgs("no-existe").
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()))

		accessKey, err := store.GetExpire(context.Background(), "no-existe")

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Equal(t, uuid.Nil, accessKey.ID)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetExpire).
			WithArgs(testAccessKeyCod).
			WillReturnError(errDBBoom)

		accessKey, err := store.GetExpire(context.Background(), testAccessKeyCod)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, uuid.Nil, accessKey.ID)
	})
}

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

	t.Run("devuelve util.ErrNotFound si no se actualizo ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.setExpire(context.Background(), expire, testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectExec(regexUpdate).
			WithArgs(expire, testUserID).
			WillReturnError(errDBBoom)

		err := store.setExpire(context.Background(), expire, testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, util.ErrNotFound)
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

func TestUserStore_GetExpire_query(t *testing.T) {
	// El Scan de GetExpire y su SELECT estan desalineados si el orden o la
	// cantidad de columnas difieren: el Scan se come un valor de la columna
	// vecina sin que ningun test lo note.
	t.Run("la proyeccion esta en el mismo orden que el Scan", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(accessKeyRow(expire)...))

		_, err := store.GetExpire(context.Background(), testAccessKeyCod)

		require.NoError(t, err)
		assert.Equal(t, accessKeyColumns(), columnasDe(*sql))
	})

	t.Run("created_at es una columna propia, no un alias de revoked_by", func(t *testing.T) {
		// "revoked_by created_at" es SQL valido en Postgres (alias implicito),
		// por eso el typo no se ve al leer la query: devuelve revoked_by dos
		// veces, created_at queda NULL y el Scan lee la columna equivocada.
		store, mock, sql := newCapturingMockStore(t)

		expire := time.Now().Add(24 * time.Hour)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(accessKeyRow(expire)...))

		_, err := store.GetExpire(context.Background(), testAccessKeyCod)

		require.NoError(t, err)
		assert.NotContains(t, *sql, "revoked_by created_at")
		assert.Contains(t, *sql, "revoked_by, created_at")
	})
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
}
