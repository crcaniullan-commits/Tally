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
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// newMockStore levanta un AlmacenUsuarios sobre una conexión sqlmockeada.
func newMockStore(t *testing.T) (*AlmacenUsuarios, sqlmock.Sqlmock) {
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

	return NuevoAlmacenUsuarios(db), mock
}

// Fragments de las queries reales, para no duplicar el SQL en cada test.
var testPlanExpiresAt = time.Date(2027, time.January, 15, 12, 0, 0, 0, time.UTC)

var (
	regexGetByID  = regexp.QuoteMeta("SELECT id, email, nombre, rut, role")
	regexGetByRut = regexp.QuoteMeta("SELECT id, email, nombre, rut, role")
	regexUpdate   = regexp.QuoteMeta("UPDATE users")
	regexDelete   = regexp.QuoteMeta("DELETE FROM users")
)

// columnas de la query de lectura: id, email, nombre, rut, role,
// plan_expires_at. El orden tiene que calzar con el Scan de ObtenerPorRUT.
func userColumns() []string {
	return []string{"id", "email", "nombre", "rut", "role", "plan_expires_at"}
}

// userColumnsConPassword son las columnas de ObtenerPorID, que además trae el
// hash de la contraseña: el middleware de autenticación usa ese usuario y
// PATCH /usuarios reescribe la fila con el hash que viene de ahí.
func userColumnsConPassword() []string {
	return []string{"id", "email", "nombre", "rut", "role", "plan_expires_at", "password_hash"}
}

// userRowConPassword es userRow() con el hash al final.
func userRowConPassword() []driver.Value {
	row := userRow()
	return append(row, []byte("hash-de-prueba"))
}

func userRow() []driver.Value {
	return []driver.Value{
		testUserID,
		"emprendedor@correo.cl",
		"Nombre Original",
		"19.234.567-K",
		string(utilidades.RolUsuario),
		testPlanExpiresAt,
	}
}

func TestUserStore_GetByID(t *testing.T) {
	t.Run("devuelve el usuario con el rut parseado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(userRowConPassword()...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, "emprendedor@correo.cl", user.Email)
		assert.Equal(t, "Nombre Original", user.Nombre)
		assert.Equal(t, utilidades.RolUsuario, user.Role)
		assert.Equal(t, utilidades.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
		require.NotNil(t, user.PlanExpiresAt)
		assert.True(t, testPlanExpiresAt.Equal(*user.PlanExpiresAt))
	})

	t.Run("devuelve el hash de la contraseña para que el PATCH no lo borre", func(t *testing.T) {
		// El middleware de autenticación arma el usuario del contexto con este
		// método y PATCH /usuarios reescribe la fila con ese usuario: si el hash
		// no viniera, la actualización escribiría un hash vacío en una columna
		// NOT NULL y el endpoint devolvía 500.
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(userRowConPassword()...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.NotEmpty(t, user.PasswordHash.GetHash())
	})

	t.Run("acepta el rut sin separadores", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := userRowConPassword()
		row[3] = "19234567K"

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(row...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, utilidades.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
	})

	t.Run("traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Nil(t, user)
	})

	t.Run("devuelve utilidades.ErrFormatoInvalido si el rut de la base no tiene formato", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := userRowConPassword()
		row[3] = "rut-corrupto"

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(row...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.ErrorIs(t, err, utilidades.ErrFormatoInvalido)
		assert.Nil(t, user)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		user, err := store.ObtenerPorID(context.Background(), testUserID)

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

		user, err := store.ObtenerPorRUT(context.Background(), "19.234.567-K")

		require.NoError(t, err)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, utilidades.RUT{Cuerpo: 19234567, DV: "K"}, user.Rut)
		require.NotNil(t, user.PlanExpiresAt)
		assert.True(t, testPlanExpiresAt.Equal(*user.PlanExpiresAt))
	})

	t.Run("traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnRows(sqlmock.NewRows(userColumns()))

		user, err := store.ObtenerPorRUT(context.Background(), "19.234.567-K")

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Nil(t, user)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnError(errDBBoom)

		user, err := store.ObtenerPorRUT(context.Background(), "19.234.567-K")

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

		err := store.Actualizar(context.Background(), user)

		require.NoError(t, err)
	})

	t.Run("devuelve utilidades.ErrorNoEncontrado si no se actualizo ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectExec(regexUpdate).
			WithArgs(user.PasswordHash.GetHash(), user.Nombre, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Actualizar(context.Background(), user)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		user := newTestUser()

		mock.ExpectExec(regexUpdate).
			WithArgs(user.PasswordHash.GetHash(), user.Nombre, user.ID).
			WillReturnError(errDBBoom)

		err := store.Actualizar(context.Background(), user)

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, utilidades.ErrorNoEncontrado, "un error de DB no es 'no encontrado'")
	})
}

func TestUserStore_Delete(t *testing.T) {
	t.Run("elimina el usuario con el id recibido", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Eliminar(context.Background(), testUserID)

		require.NoError(t, err)
	})

	t.Run("devuelve utilidades.ErrorNoEncontrado si no se elimino ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Eliminar(context.Background(), testUserID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexDelete).
			WithArgs(testUserID).
			WillReturnError(errDBBoom)

		err := store.Eliminar(context.Background(), testUserID)

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, utilidades.ErrorNoEncontrado)
	})
}

func TestUserStore_ContextCancelado(t *testing.T) {
	t.Run("no ejecuta nada si el contexto ya esta cancelado", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := NuevoAlmacenUsuarios(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = store.ObtenerPorID(ctx, testUserID)
		require.ErrorIs(t, err, context.Canceled)

		require.ErrorIs(t, store.Eliminar(ctx, testUserID), context.Canceled)
	})
}

// newCapturingMockStore levanta el store con un matcher permisivo que guarda el
// SQL realmente ejecutado. sqlmock no parsea SQL (solo compara strings), asi que
// sin esto una coma de menos o una columna cambiada de lugar pasarian los tests
// en verde y reventarian recien contra Postgres.
func newCapturingMockStore(t *testing.T) (*AlmacenUsuarios, sqlmock.Sqlmock, *string) {
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

	return NuevoAlmacenUsuarios(db), mock, &capturada
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

func TestUserStore_lectura_query(t *testing.T) {
	// Si la cantidad o el orden de columnas de la proyección difieren del Scan,
	// el Scan se come el valor de la columna vecina: sqlmock no lo detecta
	// (solo compara strings) y contra Postgres es "expected N destination
	// arguments in Scan, not M" en todas las filas. Como ObtenerPorID es lo que
	// usa el middleware de autenticación, eso tumbaba todos los endpoints bajo
	// /v1/app.
	t.Run("la proyección de ObtenerPorID está en el mismo orden que el Scan", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(userRowConPassword()...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, userColumnsConPassword(), columnasDe(*sql))
		require.NotNil(t, user.PlanExpiresAt)
		assert.True(t, testPlanExpiresAt.Equal(*user.PlanExpiresAt))
	})

	t.Run("la proyección de ObtenerPorRUT está en el mismo orden que el Scan", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(userRow()...))

		user, err := store.ObtenerPorRUT(context.Background(), "19.234.567-K")

		require.NoError(t, err)
		assert.Equal(t, userColumns(), columnasDe(*sql))
		require.NotNil(t, user.PlanExpiresAt)
		assert.True(t, testPlanExpiresAt.Equal(*user.PlanExpiresAt))
	})

	t.Run("ObtenerPorID deja PlanExpiresAt en nil si el usuario no tiene plan", func(t *testing.T) {
		// plan_expires_at es nullable a propósito (migrations/000004) y NULL es
		// "plan gratuito", o sea el caso de casi todos los usuarios. Por eso
		// PlanExpiresAt es *time.Time y no time.Time: escanear NULL en un
		// time.Time falla con "unsupported Scan, storing driver.Value type
		// <nil> into type *time.Time" y tumbaba el MiddlewareTokenAutenticacion.
		store, mock, sql := newCapturingMockStore(t)

		row := userRowConPassword()
		row[5] = nil

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(userColumnsConPassword()).AddRow(row...))

		user, err := store.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, userColumnsConPassword(), columnasDe(*sql))
		assert.Nil(t, user.PlanExpiresAt)
	})

	t.Run("ObtenerPorRUT deja PlanExpiresAt en nil si el usuario no tiene plan", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := userRow()
		row[5] = nil

		mock.ExpectQuery(regexGetByRut).
			WithArgs("19.234.567-K").
			WillReturnRows(sqlmock.NewRows(userColumns()).AddRow(row...))

		user, err := store.ObtenerPorRUT(context.Background(), "19.234.567-K")

		require.NoError(t, err)
		assert.Nil(t, user.PlanExpiresAt)
	})
}
