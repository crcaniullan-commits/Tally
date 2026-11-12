package clavesacceso

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	bodyResendOK = `{"datos":"Correo reenviado"}`
	bodyRevokeOK = `{"datos":"Código deshabilitado"}`
	bodyNotFound = `{"error":"no encontrado"}`
	body500      = `{"error":"el servidor encontró un problema"}`
)

// newTestHandler arma un HandlerClavesAcceso con un service mockeado y el logger
// de errorhandler apontando a un noop (para que los tests no ensucien la salida).
func newTestHandler(service IServiceClavesAcceso) *HandlerClavesAcceso {
	return NuevoHandlerClavesAcceso(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
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

	return r.WithContext(context.WithValue(r.Context(), utilidades.ContextoUsuario, newTestMunicipalUser()))
}

// requestWithClaveIDParam arma una request con el usuario del contexto y el URL
// param "claveID" que lee el handler vía chi.URLParam.
func requestWithClaveIDParam(method, target, claveID string) *http.Request {
	r := requestWithMunicipalUser(method, target, `{"email":"`+testEmail+`"}`)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("claveID", claveID)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

// requestWithCodeParam arma una request con el usuario del contexto y el URL
// param "code" que lee el handler vía chi.URLParam.
func requestWithCodeParam(method, target, code string) *http.Request {
	r := requestWithMunicipalUser(method, target, "")

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("codigo", code)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func decodeEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()

	var envelope map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &envelope), "respuesta no es JSON válido: %s", body)

	return envelope
}

func TestAccessKeysHandler_IssueForUser(t *testing.T) {
	t.Run("responde 201 con la llave creada envuelta en data", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		esperada := newTestAccessKey()
		service.On("EmitirParaUsuario", mock.Anything, testMunicipalID, testEmail).
			Return(esperada, nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		require.Equal(t, http.StatusCreated, w.Code)

		envelope := decodeEnvelope(t, w.Body.String())
		data, ok := envelope["datos"].(map[string]any)
		require.True(t, ok, "la respuesta debe traer un objeto en \"data\"")

		assert.Equal(t, esperada.ID.String(), data["id"])
		assert.Equal(t, esperada.Code, data["code"])
		assert.Equal(t, testMunicipalID.String(), data["created_by"])
		service.AssertExpectations(t)
	})

	t.Run("toma el emisor del contexto y el destinatario del payload", func(t *testing.T) {
		// El municipal autenticado es el que emite; el email del body es a quién
		// le llega el código. confundirlos emitiría la llave a nombre de otro.
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("EmitirParaUsuario", mock.Anything, mock.Anything, "destino@correo.cl").
			Return(newTestAccessKey(), nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"destino@correo.cl"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "EmitirParaUsuario", mock.Anything, testMunicipalID, "destino@correo.cl")
	})

	t.Run("responde 400 y no llama al servicio si el email falta", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email")
		service.AssertNotCalled(t, "EmitirParaUsuario", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el email no tiene formato", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"no-es-un-correo"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "EmitirParaUsuario", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el body no es JSON válido", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "EmitirParaUsuario", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si viene un campo desconocido", func(t *testing.T) {
		// utilidades.LeerJSON usa DisallowUnknownFields
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`","plan":"pro"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "EmitirParaUsuario", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 con cualquier error del servicio", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("EmitirParaUsuario", mock.Anything, mock.Anything, mock.Anything).
			Return(model.AccessKey{}, errStoreBoom).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestAccessKeysHandler_Resend(t *testing.T) {
	t.Run("responde 200 y reenvía el código de la llave de la URL", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Reenviar", mock.Anything, testAccessKeyID, testEmail).Return(nil).Once()

		r := requestWithClaveIDParam(http.MethodPost, "/key/AccessID/"+testAccessKeyID.String(), testAccessKeyID.String())
		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyResendOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("el destinatario del reenvío viene del payload", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Reenviar", mock.Anything, testAccessKeyID, "otro@correo.cl").Return(nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key/AccessID/x", `{"email":"otro@correo.cl"}`)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("claveID", testAccessKeyID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "Reenviar", mock.Anything, testAccessKeyID, "otro@correo.cl")
	})

	t.Run("responde 400 si el claveID de la URL no es un UUID", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithClaveIDParam(http.MethodPost, "/key/AccessID/no-es-uuid", "no-es-uuid")
		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Reenviar", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el payload es inválido", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithMunicipalUser(http.MethodPost, "/key/AccessID/x", `{"email":"nope"}`)

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("claveID", testAccessKeyID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Reenviar", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si la llave no existe", func(t *testing.T) {
		// ObtenerPorID traduce sql.ErrNoRows a utilidades.ErrorNoEncontrado justamente para que
		// este endpoint pueda contestar 404 en vez del 500 genérico.
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Reenviar", mock.Anything, testAccessKeyID, testEmail).
			Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithClaveIDParam(http.MethodPost, "/key/AccessID/"+testAccessKeyID.String(), testAccessKeyID.String())
		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Reenviar", mock.Anything, testAccessKeyID, testEmail).Return(errStoreBoom).Once()

		r := requestWithClaveIDParam(http.MethodPost, "/key/AccessID/"+testAccessKeyID.String(), testAccessKeyID.String())
		w := httptest.NewRecorder()

		handler.Reenviar(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestAccessKeysHandler_Revoke(t *testing.T) {
	t.Run("responde 200 y revoca el código de la URL", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Revocar", mock.Anything, testMunicipalID, testAccessKeyCod).Return(nil).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revocar(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyRevokeOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("quien revoca es el municipal del contexto, no el cuerpo", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Revocar", mock.Anything, testMunicipalID, mock.Anything).Return(nil).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revocar(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "Revocar", mock.Anything, testMunicipalID, testAccessKeyCod)
	})

	t.Run("responde 404 si el código no existe", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Revocar", mock.Anything, mock.Anything, mock.Anything).
			Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/no-existe", "no-existe")
		w := httptest.NewRecorder()

		handler.Revocar(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		service.On("Revocar", mock.Anything, mock.Anything, mock.Anything).
			Return(errStoreBoom).Once()

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/"+testAccessKeyCod, testAccessKeyCod)
		w := httptest.NewRecorder()

		handler.Revocar(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, body500, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 y no llama al servicio si el código viene vacío", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		r := requestWithCodeParam(http.MethodPost, "/key/revoke/", "")
		w := httptest.NewRecorder()

		handler.Revocar(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "debe ser un código")
		service.AssertNotCalled(t, "Revocar", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAccessKeysHandler_ServiceAndHandlerWiring(t *testing.T) {
	t.Run("ServicioClavesAcceso satisface la interfaz que espera el handler", func(t *testing.T) {
		// Guarda contra regresiones de firmas: si cambia IServiceClavesAcceso, esto
		// deja de compilar.
		var _ IServiceClavesAcceso = NuevoServicioClavesAcceso(new(RepositorioClavesAccesoMock), newEmailQueAcepta(nil))
	})

	t.Run("el handler construye un ServicioClavesAcceso real sobre un store mockeado", func(t *testing.T) {
		store := new(RepositorioClavesAccesoMock)
		email := newEmailQueAcepta(nil)
		store.On("Crear", mock.Anything, mock.Anything).Return(nil).Once()

		handler := newTestHandler(NuevoServicioClavesAcceso(store, email))

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		w := httptest.NewRecorder()

		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		store.AssertExpectations(t)
		email.esperar(t, 1)
	})

	t.Run("el uuid del municipal del contexto es el que se propaga", func(t *testing.T) {
		service := new(IServiceClavesAccesoMock)
		handler := newTestHandler(service)

		usuario := newTestMunicipalUser()
		service.On("EmitirParaUsuario", mock.Anything, usuario.ID, mock.Anything).
			Return(newTestAccessKey(), nil).Once()

		r := requestWithMunicipalUser(http.MethodPost, "/key", `{"email":"`+testEmail+`"}`)
		// sanity: el helper inyecta el municipal que esperamos
		require.Equal(t, usuario.ID, utilidades.ObtenerUsuarioDelContexto(r).ID)
		assert.NotEqual(t, uuid.Nil, usuario.ID)

		w := httptest.NewRecorder()
		handler.EmitirParaUsuario(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "EmitirParaUsuario", mock.Anything, uuid.MustParse(usuario.ID.String()), testEmail)
	})
}
