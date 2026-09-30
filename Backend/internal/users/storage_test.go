package users

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
