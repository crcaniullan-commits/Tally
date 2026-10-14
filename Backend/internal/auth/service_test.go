package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errServiceBoom = errors.New("boom: fallo en la capa de persistencia")

// newTxService arma un AuthService cuyo transactor corre sobre una conexion
// sqlmockeada. RegisterUser abre una transaccion de verdad, asi que el test
// tiene que poder esperar el BEGIN y el COMMIT (o el ROLLBACK) que
// dbtx.Transactor emite por debajo.
func newTxService(t *testing.T, store AuthStore, redeemer Redeemer, auth Authenticator) (*AuthService, sqlmock.Sqlmock) {
	t.Helper()

	db, sqlm, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := sqlm.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		_ = db.Close()
	})

	return NewAuthService(store, auth, dbtx.NewTransactor(db), redeemer), sqlm
}

// newPlainService arma un AuthService sin transactor: Login y CreateMunicipal
// no abren transaccion, asi que nil es seguro.
func newPlainService(store AuthStore, auth Authenticator) *AuthService {
	return NewAuthService(store, auth, nil, nil)
}

func TestAuthService_RegisterUser(t *testing.T) {
	t.Run("crea el usuario, canjea el codigo y extiende el plan en una transaccion", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		auth := new(AuthenticatorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		payload := newTestRegisterPayload()
		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Redeem", mock.Anything, testAccessKeyCode, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()
		sqlm.ExpectCommit()

		token, err := service.RegisterUser(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		store.AssertExpectations(t)
		redeemer.AssertExpectations(t)
		auth.AssertExpectations(t)
	})

	t.Run("el usuario se crea con rol usuario y el rut parseado", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		auth := new(AuthenticatorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		payload := newTestRegisterPayload()
		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Email == payload.Email &&
				u.Nombre == payload.Name &&
				u.Role == util.UserRoleUsuario &&
				u.Rut == util.RUT{Cuerpo: 19234567, DV: "K"} &&
				u.PasswordHash.Compare(payload.Password) == nil
		})).Run(func(args mock.Arguments) {
			args.Get(1).(*model.User).ID = testUserID
		}).Return(nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Redeem", mock.Anything, mock.Anything, mock.Anything).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		sqlm.ExpectCommit()

		_, err := service.RegisterUser(context.Background(), payload)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("revierte y no canjea nada si el email ya existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer, new(AuthenticatorMock))

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateEmail).Once()
		sqlm.ExpectRollback()

		token, err := service.RegisterUser(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
		assert.Empty(t, token)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("revierte y no canjea nada si el rut ya existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		service, sqlm := newTxService(t, store, redeemer, new(AuthenticatorMock))

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateRut).Once()
		sqlm.ExpectRollback()

		_, err := service.RegisterUser(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, users.ErrDuplicateRut)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("no toca el store si el rut tiene formato invalido", func(t *testing.T) {
		store := new(AuthStoreMock)
		service, sqlm := newTxService(t, store, new(RedeemerMock), new(AuthenticatorMock))

		payload := newTestRegisterPayload()
		payload.Rut = "no-es-un-rut"

		sqlm.ExpectBegin()
		sqlm.ExpectRollback()

		_, err := service.RegisterUser(context.Background(), payload)

		require.ErrorIs(t, err, util.ErrFormatoInvalido)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("no toca el store si el hasheo de password falla", func(t *testing.T) {
		store := new(AuthStoreMock)
		service, sqlm := newTxService(t, store, new(RedeemerMock), new(AuthenticatorMock))

		payload := newTestRegisterPayload()
		payload.Password = strings.Repeat("a", 73)

		sqlm.ExpectBegin()
		sqlm.ExpectRollback()

		_, err := service.RegisterUser(context.Background(), payload)

		require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("revierte y no extiende el plan si el codigo no es canjeable", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		auth := new(AuthenticatorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Redeem", mock.Anything, mock.Anything, testUserID).
			Return(nil, accesskeys.ErrNotRedeemable).Once()
		sqlm.ExpectRollback()

		_, err := service.RegisterUser(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, accesskeys.ErrNotRedeemable)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("revierte el usuario si falla la extension del plan", func(t *testing.T) {
		// Si el error de setExpire se ignorara, la transaccion commitearia y
		// quedaria un usuario creado con el codigo canjeado pero sin plan.
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		auth := new(AuthenticatorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Redeem", mock.Anything, mock.Anything, testUserID).
			Return(&accessKey, nil).Once()
		store.On("setExpire", mock.Anything, mock.Anything, testUserID).
			Return(errServiceBoom).Once()
		sqlm.ExpectRollback()

		_, err := service.RegisterUser(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, errServiceBoom)
		store.AssertExpectations(t)
	})

	t.Run("no intenta nada si el contexto ya esta cancelado", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(RedeemerMock)
		service, _ := newTxService(t, store, redeemer, new(AuthenticatorMock))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := service.RegisterUser(ctx, newTestRegisterPayload())

		require.ErrorIs(t, err, context.Canceled)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("devuelve un token cuando las credenciales son correctas", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AuthenticatorMock)
		service := newPlainService(store, auth)

		store.On("GetByEmail", mock.Anything, "emprendedor@correo.cl").
			Return(newTestUser(), nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()

		token, err := service.Login(context.Background(), &LoginUserPayload{
			Email:    "emprendedor@correo.cl",
			Password: testPassword,
		})

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		auth.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound si el email no existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AuthenticatorMock))

		store.On("GetByEmail", mock.Anything, mock.Anything).Return(nil, util.ErrNotFound).Once()

		token, err := service.Login(context.Background(), &LoginUserPayload{
			Email:    "nadie@correo.cl",
			Password: testPassword,
		})

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Empty(t, token)
	})

	t.Run("devuelve bcrypt.ErrMismatchedHashAndPassword si la clave no coincide", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AuthenticatorMock))

		store.On("GetByEmail", mock.Anything, mock.Anything).Return(newTestUser(), nil).Once()

		_, err := service.Login(context.Background(), &LoginUserPayload{
			Email:    "emprendedor@correo.cl",
			Password: "ClaveEquivocada123",
		})

		require.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
	})

	t.Run("propaga el error si generar el token falla", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AuthenticatorMock)
		service := newPlainService(store, auth)

		store.On("GetByEmail", mock.Anything, mock.Anything).Return(newTestUser(), nil).Once()
		auth.On("GenerateToken", mock.Anything).Return("", errServiceBoom).Once()

		_, err := service.Login(context.Background(), &LoginUserPayload{
			Email:    "emprendedor@correo.cl",
			Password: testPassword,
		})

		require.ErrorIs(t, err, errServiceBoom)
	})
}

func TestAuthService_CreateMunicipal(t *testing.T) {
	t.Run("crea el municipal con rol municipal sin abrir transaccion ni canjear", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AuthenticatorMock)
		service := newPlainService(store, auth)

		payload := newTestMunicipalPayload()

		store.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Email == payload.Email && u.Role == util.UserRoleMunicipal
		})).Run(func(args mock.Arguments) {
			args.Get(1).(*model.User).ID = testUserID
		}).Return(nil).Once()
		auth.On("GenerateToken", mock.Anything).Return(testToken, nil).Once()

		token, err := service.CreateMunicipal(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error de duplicado", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AuthenticatorMock))

		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateEmail).Once()

		_, err := service.CreateMunicipal(context.Background(), newTestMunicipalPayload())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
	})
}
