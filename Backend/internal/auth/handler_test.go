package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	bodyTokenOK       = `{"data":"jwt.token.de.prueba"}`
	bodyInternalError = `{"error":"the server encountered a problem"}`
)

func newTestHandler(service ServiceAuth) *AuthHandler {
	return NewAuthHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
}

func request(method, target, body string) *http.Request {
	if body == "" {
		return httptest.NewRequest(method, target, nil)
	}

	return httptest.NewRequest(method, target, strings.NewReader(body))
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("responde 200 con el token y llama al service con el payload", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("RegisterUser", mock.Anything, &CreateUserPayload{
			Email:    "nuevo@correo.cl",
			Password: testPassword,
			Name:     "Nombre Nuevo",
			Rut:      "19234567K",
			Code:     testAccessKeyCode,
		}).Return(testToken, nil).Once()

		body := `{"email":"nuevo@correo.cl","password":"ClaveSegura123","nombre":"Nombre Nuevo","rut":"19234567K","code":"` + testAccessKeyCode + `"}`

		w := httptest.NewRecorder()
		handler.RegisterHandler(w, request(http.MethodPost, "/auth", body))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyTokenOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 y no llama al service si la validacion falla", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		body := `{"email":"nuevo@correo.cl","password":"abc","nombre":"Nombre Nuevo","rut":"19234567K","code":"` + testAccessKeyCode + `"}`

		w := httptest.NewRecorder()
		handler.RegisterHandler(w, request(http.MethodPost, "/auth", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Password")
		service.AssertNotCalled(t, "RegisterUser", mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el email ya existe", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("RegisterUser", mock.Anything, mock.Anything).
			Return("", users.ErrDuplicateEmail).Once()

		body := `{"email":"nuevo@correo.cl","password":"ClaveSegura123","nombre":"Nombre Nuevo","rut":"19234567K","code":"` + testAccessKeyCode + `"}`

		w := httptest.NewRecorder()
		handler.RegisterHandler(w, request(http.MethodPost, "/auth", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"a user with that email already exists"}`, w.Body.String())
	})

	t.Run("responde 400 si el codigo no es canjeable", func(t *testing.T) {
		// El canje vive ahora en RegisterUser: un codigo invalido/vencido/
		// revocado/ya usado no puede volver a ser un 500.
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("RegisterUser", mock.Anything, mock.Anything).
			Return("", accesskeys.ErrNotRedeemable).Once()

		body := `{"email":"nuevo@correo.cl","password":"ClaveSegura123","nombre":"Nombre Nuevo","rut":"19234567K","code":"MALO"}`

		w := httptest.NewRecorder()
		handler.RegisterHandler(w, request(http.MethodPost, "/auth", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), accesskeys.ErrNotRedeemable.Error())
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("RegisterUser", mock.Anything, mock.Anything).
			Return("", errServiceBoom).Once()

		body := `{"email":"nuevo@correo.cl","password":"ClaveSegura123","nombre":"Nombre Nuevo","rut":"19234567K","code":"` + testAccessKeyCode + `"}`

		w := httptest.NewRecorder()
		handler.RegisterHandler(w, request(http.MethodPost, "/auth", body))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("responde 202 con el token", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("Login", mock.Anything, &LoginUserPayload{
			Email:    "emprendedor@correo.cl",
			Password: testPassword,
		}).Return(testToken, nil).Once()

		body := `{"email":"emprendedor@correo.cl","password":"ClaveSegura123"}`

		w := httptest.NewRecorder()
		handler.Login(w, request(http.MethodPost, "/auth/login", body))

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.JSONEq(t, bodyTokenOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si la validacion falla", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		w := httptest.NewRecorder()
		handler.Login(w, request(http.MethodPost, "/auth/login", `{"email":"no-es-email","password":"abc"}`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Login", mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el email no existe", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("Login", mock.Anything, mock.Anything).Return("", util.ErrNotFound).Once()

		body := `{"email":"nadie@correo.cl","password":"ClaveSegura123"}`

		w := httptest.NewRecorder()
		handler.Login(w, request(http.MethodPost, "/auth/login", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"resource not found"}`, w.Body.String())
	})

	t.Run("responde 400 si la clave no coincide", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("Login", mock.Anything, mock.Anything).
			Return("", bcrypt.ErrMismatchedHashAndPassword).Once()

		body := `{"email":"emprendedor@correo.cl","password":"ClaveEquivocada123"}`

		w := httptest.NewRecorder()
		handler.Login(w, request(http.MethodPost, "/auth/login", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("responde 500 con cualquier otro error", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("Login", mock.Anything, mock.Anything).Return("", errServiceBoom).Once()

		body := `{"email":"emprendedor@correo.cl","password":"ClaveSegura123"}`

		w := httptest.NewRecorder()
		handler.Login(w, request(http.MethodPost, "/auth/login", body))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
	})
}

func TestAuthHandler_CrearMunicipal(t *testing.T) {
	t.Run("responde 200 sin exigir code en el payload", func(t *testing.T) {
		// Regresion: el payload municipal compartido con registro exigia code,
		// rompiendo la creacion de municipales.
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("CreateMunicipal", mock.Anything, &CreateMunicipalPayload{
			Email:    "municipal@correo.cl",
			Password: testPassword,
			Name:     "Nombre Municipal",
			Rut:      "19234567K",
		}).Return(testToken, nil).Once()

		body := `{"email":"municipal@correo.cl","password":"ClaveSegura123","nombre":"Nombre Municipal","rut":"19234567K"}`

		w := httptest.NewRecorder()
		handler.CrearMunicipal(w, request(http.MethodPost, "/auth/municipal", body))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyTokenOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si la validacion falla", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		body := `{"email":"municipal@correo.cl","password":"abc","nombre":"Nombre Municipal","rut":"19234567K"}`

		w := httptest.NewRecorder()
		handler.CrearMunicipal(w, request(http.MethodPost, "/auth/municipal", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CreateMunicipal", mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el rut ya existe", func(t *testing.T) {
		service := new(ServiceAuthMock)
		handler := newTestHandler(service)

		service.On("CreateMunicipal", mock.Anything, mock.Anything).
			Return("", users.ErrDuplicateRut).Once()

		body := `{"email":"municipal@correo.cl","password":"ClaveSegura123","nombre":"Nombre Municipal","rut":"19234567K"}`

		w := httptest.NewRecorder()
		handler.CrearMunicipal(w, request(http.MethodPost, "/auth/municipal", body))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"a user with that rut already exists"}`, w.Body.String())
	})
}
