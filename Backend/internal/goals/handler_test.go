package goals

import (
	"context"
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

const (
	bodyGoalCreated = `{"datos":{
		"id": "d4e5f6a7-8b9c-4d0e-9f1a-2b3c4d5e6f7a",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre": "Ahorrar",
		"periodo": "mes",
		"monto_meta": 10000,
		"created_at": "2026-11-05T12:00:00Z"
	}}`
	bodyGoalState = `{"datos":{
		"monto_actual": 2500,
		"porcentaje_actual": 25,
		"monto_pendiente": 7500,
		"porcentaje_faltante": 75
	}}`
	bodyGoalList = `{"datos":[{
		"id": "d4e5f6a7-8b9c-4d0e-9f1a-2b3c4d5e6f7a",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre": "Ahorrar",
		"periodo": "mes",
		"monto_meta": 10000,
		"created_at": "2026-11-05T12:00:00Z"
	}]}`
	bodyGoalDeleted   = `{"datos":"Meta eliminada"}`
	bodyNotFound      = `{"error":"no encontrado"}`
	bodyInternalError = `{"error":"el servidor encontró un problema"}`
)

func newTestHandler(service IServiceMetas) *HandlerMetas {
	return NuevoHandlerMetas(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

func newTestEmpresaContext(r *http.Request) *http.Request {
	usuario := &model.User{ID: testUserID, Role: utilidades.RolUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, UserID: testUserID}

	ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
	ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

	return r.WithContext(ctx)
}

func requestWithEmpresa(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	return newTestEmpresaContext(r)
}

func requestWithGoalID(method, target, metaID, body string) *http.Request {
	r := requestWithEmpresa(method, target, body)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("metaID", metaID)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)

	return r.WithContext(ctx)
}

func TestGoalsHandler_Create(t *testing.T) {
	t.Run("responde 201 con el goal creado", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		payload := newTestPayloadGoal()

		service.On("CrearMeta", mock.Anything, payload, testEmpresaID).
			Return(newTestGoal(), nil).Once()

		body := `{"nombre":"Ahorrar","periodo":"mes","monto_meta":10000}`
		r := requestWithEmpresa(http.MethodPost, "/goal", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyGoalCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el JSON es invalido", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		r := requestWithEmpresa(http.MethodPost, "/goal", `{"nombre":`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CrearMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el payload no pasa la validacion", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"","periodo":"anio","monto_meta":0}`
		r := requestWithEmpresa(http.MethodPost, "/goal", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CrearMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si hay un campo desconocido", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Ahorrar","periodo":"mes","monto_meta":10000,"extra":1}`
		r := requestWithEmpresa(http.MethodPost, "/goal", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CrearMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si falta la empresa en el contexto", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Ahorrar","periodo":"mes","monto_meta":10000}`
		r := httptest.NewRequest(http.MethodPost, "/goal", strings.NewReader(body))

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CrearMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("CrearMeta", mock.Anything, mock.Anything, testEmpresaID).
			Return(model.Goal{}, errStoreBoom).Once()

		body := `{"nombre":"Ahorrar","periodo":"mes","monto_meta":10000}`
		r := requestWithEmpresa(http.MethodPost, "/goal", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestGoalsHandler_DeleteGoal(t *testing.T) {
	t.Run("responde 200 con el mensaje de confirmacion", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("EliminarMeta", mock.Anything, testGoalID, testEmpresaID).Return(nil).Once()

		r := requestWithGoalID(http.MethodDelete, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.EliminarMeta(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyGoalDeleted, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el metaID no es un UUID", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		r := requestWithGoalID(http.MethodDelete, "/goal/no-es-uuid", "no-es-uuid", "")

		w := httptest.NewRecorder()
		handler.EliminarMeta(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "EliminarMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el goal no existe", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("EliminarMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithGoalID(http.MethodDelete, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.EliminarMeta(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("EliminarMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(errStoreBoom).Once()

		r := requestWithGoalID(http.MethodDelete, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.EliminarMeta(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestGoalsHandler_GetGoalState(t *testing.T) {
	t.Run("responde 200 con el estado del goal", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		state := EstadoMeta{
			MontoActual:        2500,
			PorcentajeActual:   25,
			MontoPendiente:     7500,
			PorcentajeFaltante: 75,
		}

		service.On("ObtenerEstadoMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(state, nil).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String()+"/state", testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerEstadoMeta(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyGoalState, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el metaID no es un UUID", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		r := requestWithGoalID(http.MethodGet, "/goal/no-es-uuid/state", "no-es-uuid", "")

		w := httptest.NewRecorder()
		handler.ObtenerEstadoMeta(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "ObtenerEstadoMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el goal no existe", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerEstadoMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(EstadoMeta{}, utilidades.ErrorNoEncontrado).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String()+"/state", testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerEstadoMeta(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerEstadoMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(EstadoMeta{}, errStoreBoom).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String()+"/state", testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerEstadoMeta(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestGoalsHandler_getGoal(t *testing.T) {
	t.Run("responde 200 con el goal", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(newTestGoal(), nil).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerMeta(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyGoalCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el metaID no es un UUID", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		r := requestWithGoalID(http.MethodGet, "/goal/no-es-uuid", "no-es-uuid", "")

		w := httptest.NewRecorder()
		handler.ObtenerMeta(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "ObtenerMeta", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el goal no existe", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(model.Goal{}, utilidades.ErrorNoEncontrado).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerMeta(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMeta", mock.Anything, testGoalID, testEmpresaID).
			Return(model.Goal{}, errStoreBoom).Once()

		r := requestWithGoalID(http.MethodGet, "/goal/"+testGoalID.String(), testGoalID.String(), "")

		w := httptest.NewRecorder()
		handler.ObtenerMeta(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestGoalsHandler_getGoals(t *testing.T) {
	t.Run("responde 200 con la lista de goals", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMetas", mock.Anything, testEmpresaID).
			Return([]model.Goal{newTestGoal()}, nil).Once()

		r := requestWithEmpresa(http.MethodGet, "/goal", "")

		w := httptest.NewRecorder()
		handler.ObtenerMetas(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyGoalList, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 200 con lista vacia", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMetas", mock.Anything, testEmpresaID).
			Return([]model.Goal{}, nil).Once()

		r := requestWithEmpresa(http.MethodGet, "/goal", "")

		w := httptest.NewRecorder()
		handler.ObtenerMetas(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"datos":[]}`, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(IServiceGoalsMock)
		handler := newTestHandler(service)

		service.On("ObtenerMetas", mock.Anything, testEmpresaID).
			Return(nil, errStoreBoom).Once()

		r := requestWithEmpresa(http.MethodGet, "/goal", "")

		w := httptest.NewRecorder()
		handler.ObtenerMetas(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}
