package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	clavesacceso "github.com/crcaniullan-commits/Tally/internal/clavesAcceso"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errServiceBoom = errors.New("boom: fallo en la capa de persistencia")

// newTxService arma un ServicioAuth cuyo transactor corre sobre una conexion
// sqlmockeada. RegistrarUsuario abre una transaccion de verdad, asi que el test
// tiene que poder esperar el BEGIN y el COMMIT (o el ROLLBACK) que
// bdtx.Transactor emite por debajo.
func newTxService(t *testing.T, store RepositorioAuth, redeemer Canjeador, auth Authenticator) (*ServicioAuth, sqlmock.Sqlmock) {
	t.Helper()

	db, sqlm, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := sqlm.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		_ = db.Close()
	})

	return NuevoServicioAuth(store, auth, bdtx.NewTransactor(db), redeemer), sqlm
}

// newPlainService arma un ServicioAuth sin transactor: Login y CrearMunicipal
// no abren transaccion, asi que nil es seguro.
func newPlainService(store RepositorioAuth, auth Authenticator) *ServicioAuth {
	return NuevoServicioAuth(store, auth, nil, nil)
}

func TestAuthService_RegisterUser(t *testing.T) {
	t.Run("crea el usuario, canjea el codigo y extiende el plan en una transaccion", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		auth := new(AutenticadorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		payload := newTestRegisterPayload()
		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Canjear", mock.Anything, testAccessKeyCode, testUserID).
			Return(&accessKey, nil).Once()
		store.On("establecerVencimiento", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()
		sqlm.ExpectCommit()

		token, err := service.RegistrarUsuario(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		store.AssertExpectations(t)
		redeemer.AssertExpectations(t)
		auth.AssertExpectations(t)
	})

	t.Run("el usuario se crea con rol usuario y el rut parseado", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		auth := new(AutenticadorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		payload := newTestRegisterPayload()
		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Email == payload.Email &&
				u.Nombre == payload.Nombre &&
				u.Role == utilidades.RolUsuario &&
				u.Rut == utilidades.RUT{Cuerpo: 19234567, DV: "K"} &&
				u.PasswordHash.Compare(payload.Password) == nil
		})).Run(func(args mock.Arguments) {
			args.Get(1).(*model.User).ID = testUserID
		}).Return(nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Canjear", mock.Anything, mock.Anything, mock.Anything).
			Return(&accessKey, nil).Once()
		store.On("establecerVencimiento", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		sqlm.ExpectCommit()

		_, err := service.RegistrarUsuario(context.Background(), payload)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("revierte y no canjea nada si el email ya existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		service, sqlm := newTxService(t, store, redeemer, new(AutenticadorMock))

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateEmail).Once()
		sqlm.ExpectRollback()

		token, err := service.RegistrarUsuario(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
		assert.Empty(t, token)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
		store.AssertNotCalled(t, "establecerVencimiento", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("revierte y no canjea nada si el rut ya existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		service, sqlm := newTxService(t, store, redeemer, new(AutenticadorMock))

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateRut).Once()
		sqlm.ExpectRollback()

		_, err := service.RegistrarUsuario(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, users.ErrDuplicateRut)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("no toca el store si el rut tiene formato invalido", func(t *testing.T) {
		store := new(AuthStoreMock)
		service, sqlm := newTxService(t, store, new(CanjeadorMock), new(AutenticadorMock))

		payload := newTestRegisterPayload()
		payload.Rut = "no-es-un-rut"

		sqlm.ExpectBegin()
		sqlm.ExpectRollback()

		_, err := service.RegistrarUsuario(context.Background(), payload)

		require.ErrorIs(t, err, utilidades.ErrFormatoInvalido)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("no toca el store si el hasheo de password falla", func(t *testing.T) {
		store := new(AuthStoreMock)
		service, sqlm := newTxService(t, store, new(CanjeadorMock), new(AutenticadorMock))

		payload := newTestRegisterPayload()
		payload.Password = strings.Repeat("a", 73)

		sqlm.ExpectBegin()
		sqlm.ExpectRollback()

		_, err := service.RegistrarUsuario(context.Background(), payload)

		require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("revierte y no extiende el plan si el codigo no es canjeable", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		auth := new(AutenticadorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Canjear", mock.Anything, mock.Anything, testUserID).
			Return(nil, clavesacceso.ErrorCodigoNoCanjeable).Once()
		sqlm.ExpectRollback()

		_, err := service.RegistrarUsuario(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, clavesacceso.ErrorCodigoNoCanjeable)
		store.AssertNotCalled(t, "establecerVencimiento", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("revierte el usuario si falla la extension del plan", func(t *testing.T) {
		// Si el error de establecerVencimiento se ignorara, la transaccion commitearia y
		// quedaria un usuario creado con el codigo canjeado pero sin plan.
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		auth := new(AutenticadorMock)
		service, sqlm := newTxService(t, store, redeemer, auth)

		accessKey := newTestAccessKey()

		sqlm.ExpectBegin()
		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				args.Get(1).(*model.User).ID = testUserID
			}).Return(nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()
		redeemer.On("Canjear", mock.Anything, mock.Anything, testUserID).
			Return(&accessKey, nil).Once()
		store.On("establecerVencimiento", mock.Anything, mock.Anything, testUserID).
			Return(errServiceBoom).Once()
		sqlm.ExpectRollback()

		_, err := service.RegistrarUsuario(context.Background(), newTestRegisterPayload())

		require.ErrorIs(t, err, errServiceBoom)
		store.AssertExpectations(t)
	})

	t.Run("no intenta nada si el contexto ya esta cancelado", func(t *testing.T) {
		store := new(AuthStoreMock)
		redeemer := new(CanjeadorMock)
		service, _ := newTxService(t, store, redeemer, new(AutenticadorMock))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := service.RegistrarUsuario(ctx, newTestRegisterPayload())

		require.ErrorIs(t, err, context.Canceled)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		redeemer.AssertNotCalled(t, "Redeem", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("devuelve un token cuando las credenciales son correctas", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AutenticadorMock)
		service := newPlainService(store, auth)

		store.On("ObtenerPorEmail", mock.Anything, "emprendedor@correo.cl").
			Return(newTestUser(), nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()

		token, err := service.Login(context.Background(), &PayloadInicioSesion{
			Email:    "emprendedor@correo.cl",
			Password: testPassword,
		})

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		auth.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado si el email no existe", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AutenticadorMock))

		store.On("ObtenerPorEmail", mock.Anything, mock.Anything).Return(nil, utilidades.ErrorNoEncontrado).Once()

		token, err := service.Login(context.Background(), &PayloadInicioSesion{
			Email:    "nadie@correo.cl",
			Password: testPassword,
		})

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Empty(t, token)
	})

	t.Run("devuelve bcrypt.ErrMismatchedHashAndPassword si la clave no coincide", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AutenticadorMock))

		store.On("ObtenerPorEmail", mock.Anything, mock.Anything).Return(newTestUser(), nil).Once()

		_, err := service.Login(context.Background(), &PayloadInicioSesion{
			Email:    "emprendedor@correo.cl",
			Password: "ClaveEquivocada123",
		})

		require.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
	})

	t.Run("propaga el error si generar el token falla", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AutenticadorMock)
		service := newPlainService(store, auth)

		store.On("ObtenerPorEmail", mock.Anything, mock.Anything).Return(newTestUser(), nil).Once()
		auth.On("GenerarToken", mock.Anything).Return("", errServiceBoom).Once()

		_, err := service.Login(context.Background(), &PayloadInicioSesion{
			Email:    "emprendedor@correo.cl",
			Password: testPassword,
		})

		require.ErrorIs(t, err, errServiceBoom)
	})
}

func TestAuthService_CreateMunicipal(t *testing.T) {
	t.Run("crea el municipal con rol municipal sin abrir transaccion ni canjear", func(t *testing.T) {
		store := new(AuthStoreMock)
		auth := new(AutenticadorMock)
		service := newPlainService(store, auth)

		payload := newTestMunicipalPayload()

		store.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Email == payload.Email && u.Role == utilidades.RolMunicipal
		})).Run(func(args mock.Arguments) {
			args.Get(1).(*model.User).ID = testUserID
		}).Return(nil).Once()
		auth.On("GenerarToken", mock.Anything).Return(testToken, nil).Once()

		token, err := service.CrearMunicipal(context.Background(), payload)

		require.NoError(t, err)
		assert.Equal(t, testToken, token)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error de duplicado", func(t *testing.T) {
		store := new(AuthStoreMock)
		service := newPlainService(store, new(AutenticadorMock))

		store.On("Create", mock.Anything, mock.Anything).Return(users.ErrDuplicateEmail).Once()

		_, err := service.CrearMunicipal(context.Background(), newTestMunicipalPayload())

		require.ErrorIs(t, err, users.ErrDuplicateEmail)
	})
}
