package dispositivos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestHandler(service IServiceDispositivos) *HandlerDispositivos {
	return NuevoHandlerDispositivos(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

func requestWithUserContext(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
	ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
	return r.WithContext(ctx)
}

func requestWithTokenParam(method, target, token string, body string) *http.Request {
	r := requestWithUserContext(method, target, body)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("token", token)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	return r.WithContext(ctx)
}

func TestHandlerDispositivos_Registrar(t *testing.T) {
	t.Run("401 si no hay usuario en contexto", func(t *testing.T) {
		h := newTestHandler(new(IServiceDispositivosMock))
		r := httptest.NewRequest(http.MethodPut, "/dispositivos", strings.NewReader(`{"token":"abc","platform":"android"}`))
		w := httptest.NewRecorder()

		h.Registrar(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 si el body es inválido o faltan campos", func(t *testing.T) {
		h := newTestHandler(new(IServiceDispositivosMock))
		r := requestWithUserContext(http.MethodPut, "/dispositivos", `{"token":""}`)
		w := httptest.NewRecorder()

		h.Registrar(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 si el servicio retorna error de validación", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Registrar", mock.Anything, testUserID, PayloadDispositivo{Token: "token-invalido", Platform: "android"}).
			Return(ErrorPlataformaInvalida).Once()

		r := requestWithUserContext(http.MethodPut, "/dispositivos", `{"token":"token-invalido","platform":"android"}`)
		w := httptest.NewRecorder()

		h.Registrar(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("204 si registra exitosamente", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Registrar", mock.Anything, testUserID, PayloadDispositivo{Token: testToken, Platform: "android"}).
			Return(nil).Once()

		r := requestWithUserContext(http.MethodPut, "/dispositivos", `{"token":"`+testToken+`","platform":"android"}`)
		w := httptest.NewRecorder()

		h.Registrar(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("500 si el servicio falla internamente", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Registrar", mock.Anything, testUserID, mock.Anything).
			Return(errors.New("error inesperado")).Once()

		r := requestWithUserContext(http.MethodPut, "/dispositivos", `{"token":"`+testToken+`","platform":"android"}`)
		w := httptest.NewRecorder()

		h.Registrar(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandlerDispositivos_Desvincular(t *testing.T) {
	t.Run("401 si no hay usuario en contexto", func(t *testing.T) {
		h := newTestHandler(new(IServiceDispositivosMock))
		r := httptest.NewRequest(http.MethodDelete, "/dispositivos/abc", nil)
		w := httptest.NewRecorder()

		h.Desvincular(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 si no viene token ni en URL ni en body", func(t *testing.T) {
		h := newTestHandler(new(IServiceDispositivosMock))
		r := requestWithUserContext(http.MethodDelete, "/dispositivos", "")
		w := httptest.NewRecorder()

		h.Desvincular(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("404 si el token no existe", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Desvincular", mock.Anything, testUserID, testToken).
			Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithTokenParam(http.MethodDelete, "/dispositivos/"+testToken, testToken, "")
		w := httptest.NewRecorder()

		h.Desvincular(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("204 exitoso vía URL param", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Desvincular", mock.Anything, testUserID, testToken).
			Return(nil).Once()

		r := requestWithTokenParam(http.MethodDelete, "/dispositivos/"+testToken, testToken, "")
		w := httptest.NewRecorder()

		h.Desvincular(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("204 exitoso vía body JSON", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Desvincular", mock.Anything, testUserID, testToken).
			Return(nil).Once()

		r := requestWithUserContext(http.MethodDelete, "/dispositivos", `{"token":"`+testToken+`"}`)
		w := httptest.NewRecorder()

		h.Desvincular(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandlerDispositivos_Listar(t *testing.T) {
	t.Run("401 si no hay usuario en contexto", func(t *testing.T) {
		h := newTestHandler(new(IServiceDispositivosMock))
		r := httptest.NewRequest(http.MethodGet, "/dispositivos", nil)
		w := httptest.NewRecorder()

		h.Listar(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("200 con lista de dispositivos", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		dispositivos := []model.DeviceToken{
			{
				ID:         testDeviceID,
				UserID:     testUserID,
				Token:      testToken,
				Platform:   "android",
				LastSeenAt: testNow,
			},
		}

		svc.On("Listar", mock.Anything, testUserID).Return(dispositivos, nil).Once()

		r := requestWithUserContext(http.MethodGet, "/dispositivos", "")
		w := httptest.NewRecorder()

		h.Listar(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testToken)
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := new(IServiceDispositivosMock)
		h := newTestHandler(svc)

		svc.On("Listar", mock.Anything, testUserID).Return(nil, errors.New("fallo")).Once()

		r := requestWithUserContext(http.MethodGet, "/dispositivos", "")
		w := httptest.NewRecorder()

		h.Listar(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}
