package users

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestUserService_Update(t *testing.T) {
	t.Run("hashea la contraseña y persiste el usuario cuando viene password", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

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
		service := NewUserService(store)

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
		service := NewUserService(store)

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
		service := NewUserService(store)

		user := newTestUser()
		payload := &UpdateUserPayload{}

		store.On("Update", mock.Anything, user).Return(nil).Once()

		require.NoError(t, service.Update(context.Background(), user, payload))
		assert.Equal(t, "Nombre Original", user.Nombre)
		store.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

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
		service := NewUserService(store)

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
		service := NewUserService(store)

		userID := uuid.New()
		store.On("Delete", mock.Anything, userID).Return(nil).Once()

		err := service.Delete(context.Background(), userID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "Delete", mock.Anything, userID)
	})

	t.Run("propaga util.ErrNotFound cuando el usuario no existe", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		userID := uuid.New()
		store.On("Delete", mock.Anything, userID).Return(util.ErrNotFound).Once()

		err := service.Delete(context.Background(), userID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

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
		service := NewUserService(store)

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
				service := NewUserService(store)

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
		service := NewUserService(store)

		store.On("GetByRut", mock.Anything, mock.Anything).Return(nil, util.ErrNotFound).Once()

		user, err := service.GetByRut(context.Background(), util.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.Nil(t, user)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva un usuario", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		store.On("GetByRut", mock.Anything, mock.Anything).Return(newTestUser(), errStoreBoom).Once()

		user, err := service.GetByRut(context.Background(), util.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, user, "si hay error el servicio debe devolver nil")
	})
}

func TestUserService_ExchangeCode(t *testing.T) {
	t.Run("canjea un codigo vigente y extiende el plan hasta su vencimiento", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).Return(nil).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("rechaza un codigo que ya fue canjeado", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()
		canjeado := time.Now().Add(-time.Hour)
		accessKey.RedeemedAt = &canjeado
		accessKey.RedeemedBy = &testUserID

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, ErrCodeRedemed)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("rechaza un codigo revocado aunque siga vigente", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()
		revocado := time.Now().Add(-time.Hour)
		accessKey.RevokedAt = &revocado
		accessKey.RevokedBy = &testMunicipalID

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, ErrCodeRevocado)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("rechaza un codigo vencido", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()
		vencido := time.Now().Add(-time.Minute)
		accessKey.ExpiresAt = &vencido

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, ErrCodeVencido)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("rechaza un codigo sin vencimiento en vez de entrar en panic", func(t *testing.T) {
		// expires_at es nullable (migrations/000001). Desreferenciarlo a ciegas
		// hace que el service entre en panic y el Recoverer de chi lo convierta
		// en un 500 sin cuerpo.
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()
		accessKey.ExpiresAt = nil

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, ErrCodeVencido)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga util.ErrNotFound sin tocar el store si el codigo no existe", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		store.On("GetExpire", mock.Anything, "no-existe").
			Return(model.AccessKey{}, util.ErrNotFound).Once()

		err := service.ExchangeCode(context.Background(), "no-existe", testUserID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertNotCalled(t, "setExpire", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga el error del store al extender el plan", func(t *testing.T) {
		store := new(StoreUserMock)
		service := NewUserService(store)

		accessKey := newTestAccessKey()

		store.On("GetExpire", mock.Anything, testAccessKeyCod).Return(accessKey, nil).Once()
		store.On("setExpire", mock.Anything, *accessKey.ExpiresAt, testUserID).
			Return(errStoreBoom).Once()

		err := service.ExchangeCode(context.Background(), testAccessKeyCod, testUserID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}
