package users

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

// newService arma un UserService con el store y el redeemer mockeados. Los
// metodos que no abren transaccion (Update, Delete, GetByRut) no llegan a tocar
// ni al redeemer ni al transactor, asi que nil en los dos ultimos argumentos
// es seguro y deja claro en cada test que no intervienen.
func newService(store StoreUser, redeemer Redeemer) *UserService {
	return NewUserService(store, redeemer, nil)
}

// newTxService arma un UserService cuyo transactor corre sobre una conexion
// sqlmockeada. ExchangeCode abre una transacion de verdad, asi que el test tiene
// que poder esperar el BEGIN y el COMMIT (o el ROLLBACK) que dbtx.Transactor
// emite por debajo.
func newTxService(t *testing.T, store StoreUser, redeemer Redeemer) (*UserService, sqlmock.Sqlmock) {
	t.Helper()

	db, sqlm, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := sqlm.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		_ = db.Close()
	})

	return NewUserService(store, redeemer, dbtx.NewTransactor(db)), sqlm
}

func TestUserService_Update(t *testing.T) {
	t.Run("hashea la contraseña y persiste el usuario cuando viene password", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		payload := &UpdateUserPayload{Password: "NuevaClave123", Name: "Nombre Nuevo"}

		store.On("Update", mock.Anything, user).Return(nil).Once()

		err := service.Update(context.Background(), user, payload)

		require.NoError(t, err)

		// La contraseña debe quedar hasheada y ser verificable.
		assert.NotEmpty(t, user.PasswordHash.GetHash())
		assert.NotEqual(t, []byte("NuevaClave123"), user.PasswordHash.GetHash())
		assert.NoError(t, user.PasswordHash.Compare("NuevaClave123"))

		store.AssertExpectations(t)
		store.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("actualiza nombre y contraseña a la vez cuando viene el payload completo", func(t *testing.T) {
		// Regresión: antes service.go usaba `if password != "" { ... } else if name != ""`,
		// y enviar ambos campos dejaba el nombre sin actualizar. Ahora son dos `if`
		// independientes (service.go:26 y service.go:32).
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		payload := &UpdateUserPayload{Password: "NuevaClave123", Name: "Nombre Nuevo"}

		store.On("Update", mock.Anything, user).Return(nil).Once()

		require.NoError(t, service.Update(context.Background(), user, payload))

		assert.Equal(t, "Nombre Nuevo", user.Nombre, "el nombre debe actualizarse aunque venga password")
		assert.NoError(t, user.PasswordHash.Compare("NuevaClave123"))
		store.AssertExpectations(t)
	})

	t.Run("actualiza solo el nombre cuando no viene password", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		hashPrevio := append([]byte(nil), user.PasswordHash.GetHash()...)
		payload := &UpdateUserPayload{Name: "Nombre Nuevo"}

		store.On("Update", mock.Anything, user).Return(nil).Once()

		require.NoError(t, service.Update(context.Background(), user, payload))

		assert.Equal(t, "Nombre Nuevo", user.Nombre)
		assert.Equal(t, hashPrevio, user.PasswordHash.GetHash(), "no debe tocar la contraseña")
		store.AssertExpectations(t)
	})

	t.Run("persiste aunque el payload venga vacío", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		payload := &UpdateUserPayload{}

		store.On("Update", mock.Anything, user).Return(nil).Once()

		require.NoError(t, service.Update(context.Background(), user, payload))
		assert.Equal(t, "Nombre Original", user.Nombre)
		store.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		payload := &UpdateUserPayload{Name: "Nombre Nuevo"}

		store.On("Update", mock.Anything, user).Return(util.ErrNotFound).Once()

		err := service.Update(context.Background(), user, payload)

		require.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("no toca el store si el hasheo falla (password > 72 bytes)", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		user := newTestUser()
		payload := &UpdateUserPayload{Password: strings.Repeat("a", 73)}

		err := service.Update(context.Background(), user, payload)

		require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		store.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	t.Run("elimina el usuario con el id recibido", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		userID := uuid.New()
		store.On("Delete", mock.Anything, userID).Return(nil).Once()

		err := service.Delete(context.Background(), userID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "Delete", mock.Anything, userID)
	})

	t.Run("propaga util.ErrNotFound cuando el usuario no existe", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		userID := uuid.New()
		store.On("Delete", mock.Anything, userID).Return(util.ErrNotFound).Once()

		err := service.Delete(context.Background(), userID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		userID := uuid.New()
		store.On("Delete", mock.Anything, userID).Return(errStoreBoom).Once()

		err := service.Delete(context.Background(), userID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestUserService_GetByRut(t *testing.T) {
	t.Run("busca por el RUT formateado y devuelve el usuario", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		esperado := newTestUser()
		store.On("GetByRut", mock.Anything, "19.234.567-K").Return(esperado, nil).Once()

		user, err := service.GetByRut(context.Background(), util.RUT{Cuerpo: 19234567, DV: "K"})

		require.NoError(t, err)
		assert.Same(t, esperado, user)
		store.AssertExpectations(t)
	})

	t.Run("convierte el RUT a string estándar antes de consultar el store", func(t *testing.T) {
		casos := []struct {
			nombre   string
			rut      util.RUT
			formatea string
		}{
			{"cuerpo de 7 dígitos", util.RUT{Cuerpo: 19234567, DV: "K"}, "19.234.567-K"},
			{"cuerpo de 4 dígitos", util.RUT{Cuerpo: 1234, DV: "5"}, "1.234-5"},
			{"cuerpo de 2 dígitos", util.RUT{Cuerpo: 12, DV: "3"}, "12-3"},
		}

		for _, c := range casos {
			t.Run(c.nombre, func(t *testing.T) {
				store := new(StoreUserMock)
				service := newService(store, new(RedeemerMock))

				esperado := newTestUser()
				store.On("GetByRut", mock.Anything, c.formatea).Return(esperado, nil).Once()

				user, err := service.GetByRut(context.Background(), c.rut)

				require.NoError(t, err)
				assert.Same(t, esperado, user)
				store.AssertExpectations(t)
			})
		}
	})

	t.Run("devuelve nil usuario y el error si el store falla", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		store.On("GetByRut", mock.Anything, mock.Anything).Return(nil, util.ErrNotFound).Once()

		user, err := service.GetByRut(context.Background(), util.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Nil(t, user)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva un usuario", func(t *testing.T) {
		store := new(StoreUserMock)
		service := newService(store, new(RedeemerMock))

		store.On("GetByRut", mock.Anything, mock.Anything).Return(newTestUser(), errStoreBoom).Once()

		user, err := service.GetByRut(context.Background(), util.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, user, "si hay error el servicio debe devolver nil")
	})
}

func TestUserService_ExchangeCode(t *testing.T) {
	// El canje ya no lo decide este service: el UPDATE de access_keys filtra
	// solo las llaves canjeables (redeemed_by IS NULL AND expires_at > now() AND
	// revoked_at IS NULL) y devuelve ErrNotRedeemable cuando no matchea ninguna
	// fila. O sea, "ya canjeado", "vencido" y "revocado" son el mismo error
	// desde aca, y el unico trabajo del service es correr el canje y la
	// extension del plan dentro de la misma transaccion.
	t.Run("canjea un codigo vigente y extiende el plan hasta su vencimiento", func(t *testing.T) {
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()

		redeemer.On("Redeem", mock.Anything, testAccessKeyCod, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()

		sqlm.ExpectCommit()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.NoError(t, err)
		redeemer.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("el canje y la extension del plan van en la misma transaccion", func(t *testing.T) {
		// Si el plan no se extendiera dentro de la transaccion del canje, una
		// llave couldueada con exito dejaria al usuario sin plan: el BEGIN y el
		// COMMIT de abajo solo se cumplen si dbtx.Transactor envuelve a ambos.
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		redeemer.On("Redeem", mock.Anything, testAccessKeyCod, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()
		sqlm.ExpectCommit()

		require.NoError(t, service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID))
		store.AssertCalled(t, "setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID)
	})

	t.Run("revierte y no extiende el plan si el codigo no es canjeable", func(t *testing.T) {
		// ErrNotRedeemable es lo que devuelve el store de accessKeys cuando el
		// UPDATE no matchea ninguna fila: codigo ya canjeado, vencido o revocado.
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		sqlm.ExpectBegin()
		redeemer.On("Redeem", mock.Anything, testAccessKeyCod, testUserID).
			Return(nil, accesskeys.ErrNotRedeemable).Once()
		sqlm.ExpectRollback()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, accesskeys.ErrNotRedeemable)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
		redeemer.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound del canje sin extender el plan", func(t *testing.T) {
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		sqlm.ExpectBegin()
		redeemer.On("Redeem", mock.Anything, "no-existe", testUserID).
			Return(nil, util.ErrNotFound).Once()
		sqlm.ExpectRollback()

		err := service.ExchangeCode(context.Background(), "no-existe", testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("revierte y propaga el error si extender el plan falla", func(t *testing.T) {
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		redeemer.On("Redeem", mock.Anything, testAccessKeyCod, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).
			Return(errStoreBoom).Once()
		sqlm.ExpectRollback()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del commit: la llave queda canjeada pero el plan no", func(t *testing.T) {
		// El canje y el plan viajan en la misma transaccion justamente para esto:
		// si el COMMIT falla, no hay ni llave canjeada ni plan extendido.
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		redeemer.On("Redeem", mock.Anything, testAccessKeyCod, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()
		sqlm.ExpectCommit().WillReturnError(errStoreBoom)

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})

	t.Run("no intenta canjear si el contexto ya esta cancelado", func(t *testing.T) {
		store := new(StoreUserMock)
		redeemer := new(RedeemerMock)
		service, _ := newTxService(t, store, redeemer)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := service.ExchangeCode(ctx, testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, context.Canceled)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})
}
