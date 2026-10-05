package accesskeys

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	bodyResendOK = `{"data":"Correo reenviado"}`
	bodyRevokeOK = `{"data":"Codigo desabilitado"}`
	bodyNotFound = `{"error":"not found"}`
	body500      = `{"error":"the server encountered a problem"}`
)

// newTestHandler arma un AccessKeysHandler con un service mockeado y el logger
// de errorhandler apontando a un noop (para que los tests no ensucien la salida).
func newTestHandler(service ServiceAccessKey) *AccessKeysHandler {
	return NewAccessKeysHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
}

// requestWithMunicipalUser construye una request con el usuario del middleware
// inyectado en el contexto, igual que lo hace AuthTokenMiddleware.
func requestWithMunicipalUser(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	return r.WithContext(context.WithValue(r.Context(), util.UserCtx, newTestMunicipalUser()))
}

// requestWithKeyIDParam arma una request con el usuario del contexto y el URL
// param "keyID" que lee el handler vía chi.URLParam.
func requestWithKeyIDParam(method, target, keyID string) *http.Request {
	r := requestWithMunicipalUser(method, target, `{"email":"`+testEmail+`"}`)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("keyID", keyID)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

// requestWithCodeParam arma una request con el usuario del contexto y el URL
// param "code" que lee el handler vía chi.URLParam.
func requestWithCodeParam(method, target, code string) *http.Request {
	r := requestWithMunicipalUser(method, target, "")

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func decodeEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()

	var envelope map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &envelope), "respuesta no es JSON válido: %s", body)

	return envelope
}

func TestAccessKeysHandler_IssueForUser(t *testing.T) {
	t.Run("responde 200 con la llave creada envuelta en data", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		esperada := newTestAccessKey()
		service.On("IssueForUser", mock.Anything, testMunicipalID, testEmail).
			Return(esperada, nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)

		envelope := decodeEnvelope(t, w.Body.String())
		data, ok := envelope["data"].(map[string]any)
		require.True(t, ok, "la respuesta debe traer un objeto en \"data\"")

		assert.Equal(t, esperada.ID.String(), data["id"])
		assert.Equal(t, esperada.Code, data["code"])
		assert.Equal(t, testMunicipalID.String(), data["created_by"])
		service.AssertExpectations(t)
	})

	t.Run("toma el emisor del contexto y el destinatario del payload", func(t *testing.T) {
		// El municipal autenticado es el que emite; el email del body es a quién
		// le llega el código. confundirlos emitiría la llave a nombre de otro.
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("IssueForUser", mock.Anything, mock.Anything, "destino@correo.cl").
			Return(newTestAccessKey(), nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"destino@correo.cl"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "IssueForUser", mock.Anything, testMunicipalID, "destino@correo.cl")
	})

	t.Run("responde 400 y no llama al servicio si el email falta", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email")
		service.AssertNotCalled(t, "IssueForUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el email no tiene formato", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"no-es-un-correo"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "IssueForUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el body no es JSON válido", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "IssueForUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si viene un campo desconocido", func(t *testing.T) {
		// util.ReadJSON usa DisallowUnknownFields
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`","plan":"pro"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "IssueForUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 con cualquier error del servicio", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("IssueForUser", mock.Anything, mock.Anything, mock.Anything).
			Return(model.AccessKey{}, errStoreBoom).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestAccessKeysHandler_Resend(t *testing.T) {
	t.Run("responde 200 y reenvía el código de la llave de la URL", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("Resend", mock.Anything, testAccessKeyID, testEmail).Return(nil).Once()

		r := requestWithKeyIDParam(http.MethodPost, "/key/AccessID/"+testAccessKeyID.String(), testAccessKeyID.String())
		w := httptest.NewRecorder()

		handler.Resend(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyResendOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("el destinatario del reenvío viene del payload", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("Resend", mock.Anything, testAccessKeyID, "otro@correo.cl").Return(nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key/AccessID/x", `{"email":"otro@correo.cl"}`)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("keyID", testAccessKeyID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

		w := httptest.NewRecorder()

		handler.Resend(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "Resend", mock.Anything, testAccessKeyID, "otro@correo.cl")
	})

	t.Run("responde 400 si el keyID de la URL no es un UUID", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithKeyIDParam(http.MethodPost, "/key/AccessID/no-es-uuid", "no-es-uuid")
		w := httptest.NewRecorder()

		handler.Resend(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Resend", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el payload es inválido", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key/AccessID/x", `{"email":"nope"}`)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("keyID", testAccessKeyID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

		w := httptest.NewRecorder()

		handler.Resend(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Resend", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 si la llave no existe", func(t *testing.T) {
		// GetByID no traduce sql.ErrNoRows a util.ErrNotFound, así que el
		// handler no tiene con qué responder 404 y todo error del service cae
		// en el 500 genérico.
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("Resend", mock.Anything, testAccessKeyID, testEmail).Return(errStoreBoom).Once()

		r := requestWithKeyIDParam(http.MethodPost, "/key/AccessID/"+testAccessKeyID.String(), testAccessKeyID.String())
		w := httptest.NewRecorder()

		handler.Resend(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestAccessKeysHandler_Revoke(t *testing.T) {
	t.Run("responde 200 y revoca el código de la URL", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("RevokePremature", mock.Anything, testMunicipalID, testAccessKeyCod).Return(nil).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revoke(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyRevokeOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("quien revoca es el municipal del contexto, no el cuerpo", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("RevokePremature", mock.Anything, testMunicipalID, mock.Anything).Return(nil).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revoke(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "RevokePremature", mock.Anything, testMunicipalID, testAccessKeyCod)
	})

	t.Run("responde 404 si el código no existe", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("RevokePremature", mock.Anything, mock.Anything, mock.Anything).
			Return(util.ErrNotFound).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/no-existe", "no-existe")
		w := httptest.NewRecorder()

		handler.Revoke(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("RevokePremature", mock.Anything, mock.Anything, mock.Anything).
			Return(errStoreBoom).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revoke(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("BUG: con código vacío responde 400 pero igual revoca con código vacío", func(t *testing.T) {
		// handler.go:96-98 escribe el 400 pero NO hace return, así que sigue
		// revocando con la cadena vacía y, si el service no falla, después
		// intenta escribir un 200 encima del 400 ya enviado.
		// Este test fija el comportamiento ACTUAL para hacerlo visible.
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		service.On("RevokePremature", mock.Anything, mock.Anything, "").Return(util.ErrNotFound).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/", "")
		w := httptest.NewRecorder()

		handler.Revoke(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "have to be a code")
		service.AssertCalled(t, "RevokePremature", mock.Anything, mock.Anything, "")
		service.AssertExpectations(t)
	})
}

func TestAccessKeysHandler_ServiceAndHandlerWiring(t *testing.T) {
	t.Run("AccessKeyService satisface la interfaz que espera el handler", func(t *testing.T) {
		// Guarda contra regresiones de firmas: si cambia ServiceAccessKey, esto
		// deja de compilar.
		var _ ServiceAccessKey = NewAccessKeyService(new(AccessKeyStoreMock), newEmailQueAcepta(nil))
	})

	t.Run("el handler construye un AccessKeyService real sobre un store mockeado", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		store.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		handler := newTestHandler(NewAccessKeyService(store, email))

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		store.AssertExpectations(t)
		email.esperar(t, 1)
	})

	t.Run("el uuid del municipal del contexto es el que se propaga", func(t *testing.T) {
		service := new(ServiceAccessKeyMock)
		handler := newTestHandler(service)

		usuario := newTestMunicipalUser()
		service.On("IssueForUser", mock.Anything, usuario.ID, mock.Anything).
			Return(newTestAccessKey(), nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		// sanity: el helper inyecta el municipal que esperamos
		require.Equal(t, usuario.ID, util.GetUserFromContext(r).ID)
		assert.NotEqual(t, uuid.Nil, usuario.ID)

		w := httptest.NewRecorder()
		handler.IssueForUser(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "IssueForUser", mock.Anything, uuid.MustParse(usuario.ID.String()), testEmail)
	})
}
