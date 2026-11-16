package users

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

// newService arma un ServicioUsuarios con el store mockeado. El canje de codigos ya
// no vive en este service (se movio al registro en auth), asi que no hay
// transactor ni redeemer que inyectar.
func newService(store RepositorioUsuarios) *ServicioUsuarios {
	return newServiceWithInvalidator(store, newInvalidadorUsuarioMock())
}

// newServiceWithInvalidator arma el service con un invalidador de cache
// inyectado, para poder verificar que el service lo llama.
func newServiceWithInvalidator(store RepositorioUsuarios, inv invalidarUsuario) *ServicioUsuarios {
	return NuevoServicioUsuarios(store, inv)
}

func TestUserService_Update(t *testing.T) {
	t.Run("hashea la contraseña y persiste el usuario cuando viene password", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Password: "NuevaClave123", Nombre: "Nombre Nuevo"}

		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).Return(nil).Once()

		err := service.Actualizar(context.Background(), user, payload)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertNumberOfCalls(t, "Actualizar", 1)
	})

	t.Run("el hash que manda al store es bcrypt del password nuevo", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Password: "NuevaClave123"}

		var cambios CambiosUsuario
		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).
			Run(func(args mock.Arguments) { cambios = args.Get(1).(CambiosUsuario) }).
			Return(nil).Once()

		require.NoError(t, service.Actualizar(context.Background(), user, payload))

		require.True(t, cambios.CambiaPassword)
		assert.NotEmpty(t, cambios.PasswordHash)
		assert.NotEqual(t, []byte("NuevaClave123"), cambios.PasswordHash, "el hash no es el password en claro")

		var auxiliar model.User
		auxiliar.PasswordHash.Hash = cambios.PasswordHash
		assert.NoError(t, auxiliar.PasswordHash.Compare("NuevaClave123"))
	})

	t.Run("actualiza nombre y contraseña a la vez cuando viene el payload completo", func(t *testing.T) {
		// Regresión: antes service.go usaba `if password != "" { ... } else if name != ""`,
		// y enviar ambos campos dejaba el nombre sin actualizar. Ahora son dos `if`
		// independientes.
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Password: "NuevaClave123", Nombre: "Nombre Nuevo"}

		var cambios CambiosUsuario
		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).
			Run(func(args mock.Arguments) { cambios = args.Get(1).(CambiosUsuario) }).
			Return(nil).Once()

		require.NoError(t, service.Actualizar(context.Background(), user, payload))

		require.NotNil(t, cambios.Nombre)
		assert.Equal(t, "Nombre Nuevo", *cambios.Nombre, "el nombre debe actualizarse aunque venga password")
		assert.True(t, cambios.CambiaPassword)
		store.AssertExpectations(t)
	})

	t.Run("actualiza solo el nombre cuando no viene password", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Nombre: "Nombre Nuevo"}

		var cambios CambiosUsuario
		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).
			Run(func(args mock.Arguments) { cambios = args.Get(1).(CambiosUsuario) }).
			Return(nil).Once()

		require.NoError(t, service.Actualizar(context.Background(), user, payload))

		require.NotNil(t, cambios.Nombre)
		assert.Equal(t, "Nombre Nuevo", *cambios.Nombre)
		assert.False(t, cambios.CambiaPassword, "no debe tocar la contraseña")
		assert.Empty(t, cambios.PasswordHash)
		store.AssertExpectations(t)
	})

	t.Run("no manda el hash del usuario del contexto, solo el del payload", func(t *testing.T) {
		// El usuario del contexto puede venir del cache, y ahí el hash llega
		// vacío (json:"-"). Si el service lo reenviara, el UPDATE mandaría un
		// password_hash vacío a una columna NOT NULL.
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		user.PasswordHash.Hash = nil
		payload := &PayloadActualizarUsuario{Nombre: "Nombre Nuevo"}

		var cambios CambiosUsuario
		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).
			Run(func(args mock.Arguments) { cambios = args.Get(1).(CambiosUsuario) }).
			Return(nil).Once()

		require.NoError(t, service.Actualizar(context.Background(), user, payload))

		assert.False(t, cambios.CambiaPassword)
		assert.Empty(t, cambios.PasswordHash, "no debe mandar un hash vacío")
		store.AssertExpectations(t)
	})

	t.Run("no toca el store si el payload viene vacío", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{}

		require.NoError(t, service.Actualizar(context.Background(), user, payload))
		store.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Nombre: "Nombre Nuevo"}

		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.Actualizar(context.Background(), user, payload)

		require.Error(t, err)
		assert.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("no toca el store si el hasheo falla (password > 72 bytes)", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Password: strings.Repeat("a", 73)}

		err := service.Actualizar(context.Background(), user, payload)

		require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		store.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("invalida la entrada cacheada del usuario", func(t *testing.T) {
		// Sin invalidar, el middleware seguiría sirviendo el usuario viejo
		// (por ejemplo, con el nombre anterior).
		store := new(RepositorioUsuariosMock)
		inv := new(InvalidadorUsuarioMock)
		service := newServiceWithInvalidator(store, inv)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Nombre: "Nombre Nuevo"}

		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).Return(nil).Once()
		inv.On("Invalidar", mock.Anything, user.ID).Once()

		require.NoError(t, service.Actualizar(context.Background(), user, payload))

		store.AssertExpectations(t)
		inv.AssertExpectations(t)
	})

	t.Run("no invalida si el store falla", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		inv := new(InvalidadorUsuarioMock)
		service := newServiceWithInvalidator(store, inv)

		user := newTestUser()
		payload := &PayloadActualizarUsuario{Nombre: "Nombre Nuevo"}

		store.On("Actualizar", mock.Anything, mock.Anything, user.ID).Return(errStoreBoom).Once()

		require.ErrorIs(t, service.Actualizar(context.Background(), user, payload), errStoreBoom)
		inv.AssertNotCalled(t, "Invalidar", mock.Anything, mock.Anything)
	})
}

func TestUserService_Delete(t *testing.T) {
	t.Run("elimina el usuario con el id recibido", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		userID := uuid.New()
		store.On("Eliminar", mock.Anything, userID).Return(nil).Once()

		err := service.Eliminar(context.Background(), userID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "Eliminar", mock.Anything, userID)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando el usuario no existe", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		userID := uuid.New()
		store.On("Eliminar", mock.Anything, userID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.Eliminar(context.Background(), userID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		userID := uuid.New()
		store.On("Eliminar", mock.Anything, userID).Return(errStoreBoom).Once()

		err := service.Eliminar(context.Background(), userID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})

	t.Run("invalida la entrada cacheada del usuario", func(t *testing.T) {
		// Si la entrada cacheada sobrevive al borrado, el middleware de
		// autenticación seguiría aceptando el token de una cuenta borrada.
		store := new(RepositorioUsuariosMock)
		inv := new(InvalidadorUsuarioMock)
		service := newServiceWithInvalidator(store, inv)

		userID := uuid.New()
		store.On("Eliminar", mock.Anything, userID).Return(nil).Once()
		inv.On("Invalidar", mock.Anything, userID).Once()

		err := service.Eliminar(context.Background(), userID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		inv.AssertExpectations(t)
	})

	t.Run("no invalida si el store falla", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		inv := new(InvalidadorUsuarioMock)
		service := newServiceWithInvalidator(store, inv)

		userID := uuid.New()
		store.On("Eliminar", mock.Anything, userID).Return(errStoreBoom).Once()

		err := service.Eliminar(context.Background(), userID)

		require.ErrorIs(t, err, errStoreBoom)
		inv.AssertNotCalled(t, "Invalidar", mock.Anything, mock.Anything)
	})
}

func TestUserService_GetByRut(t *testing.T) {
	t.Run("busca por el RUT formateado y devuelve el usuario", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		esperado := newTestUser()
		store.On("ObtenerPorRUT", mock.Anything, "19.234.567-K").Return(esperado, nil).Once()

		user, err := service.ObtenerPorRUT(context.Background(), utilidades.RUT{Cuerpo: 19234567, DV: "K"})

		require.NoError(t, err)
		assert.Same(t, esperado, user)
		store.AssertExpectations(t)
	})

	t.Run("convierte el RUT a string estándar antes de consultar el store", func(t *testing.T) {
		casos := []struct {
			nombre   string
			rut      utilidades.RUT
			formatea string
		}{
			{"cuerpo de 7 dígitos", utilidades.RUT{Cuerpo: 19234567, DV: "K"}, "19.234.567-K"},
			{"cuerpo de 4 dígitos", utilidades.RUT{Cuerpo: 1234, DV: "5"}, "1.234-5"},
			{"cuerpo de 2 dígitos", utilidades.RUT{Cuerpo: 12, DV: "3"}, "12-3"},
		}

		for _, c := range casos {
			t.Run(c.nombre, func(t *testing.T) {
				store := new(RepositorioUsuariosMock)
				service := newService(store)

				esperado := newTestUser()
				store.On("ObtenerPorRUT", mock.Anything, c.formatea).Return(esperado, nil).Once()

				user, err := service.ObtenerPorRUT(context.Background(), c.rut)

				require.NoError(t, err)
				assert.Same(t, esperado, user)
				store.AssertExpectations(t)
			})
		}
	})

	t.Run("devuelve nil usuario y el error si el store falla", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		store.On("ObtenerPorRUT", mock.Anything, mock.Anything).Return(nil, utilidades.ErrorNoEncontrado).Once()

		user, err := service.ObtenerPorRUT(context.Background(), utilidades.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		assert.Nil(t, user)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva un usuario", func(t *testing.T) {
		store := new(RepositorioUsuariosMock)
		service := newService(store)

		store.On("ObtenerPorRUT", mock.Anything, mock.Anything).Return(newTestUser(), errStoreBoom).Once()

		user, err := service.ObtenerPorRUT(context.Background(), utilidades.RUT{Cuerpo: 19234567, DV: "K"})

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, user, "si hay error el servicio debe devolver nil")
	})
}
