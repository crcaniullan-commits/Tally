package empresas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/crcaniullan-commits/Tally/internal/errores"
)

const (
	bodyEmpresaCreated = `{"datos":{
		"id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre_empresa": "Mi Empresa",
		"user_id": "6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b"
	}}`
	bodyEmpresasList   = `{"datos":[{"id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d","nombre_empresa": "Mi Empresa","user_id": "6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b"}]}`
	bodyNombreUpdated  = `{"datos":"Nombre actualizado"}`
	bodyEmpresaDeleted = `{"datos":"empresa borrada"}`
	bodyNotFound       = `{"error":"no encontrado"}`
	bodyInternalError  = `{"error":"el servidor encontró un problema"}`
)

// newTestHandler arma un HandlerEmpresas con un service mockeado y el logger de
// errorhandler apuntando a un noop.
func newTestHandler(service IServiceEmpresas) *HandlerEmpresas {
	return NuevoHandlerEmpresas(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

// requestWithUser construye una request con el usuario del middleware inyectado
// en el contexto.
func requestWithUser(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}

	return r.WithContext(context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario))
}

// requestWithEmpresaContext inyecta usuario y empresa en el contexto.
func requestWithEmpresaContext(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}

	ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
	ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

	return r.WithContext(ctx)
}

// requestWithNombreParam inyecta el URL param "nombre" que lee el handler
// vía chi.URLParam.
func requestWithNombreParam(method, target, nombre string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("nombre", nombre)

	r := httptest.NewRequest(method, "/test", nil)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
	ctx = context.WithValue(ctx, utilidades.ContextoUsuario, usuario)
	ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

	return r.WithContext(ctx)
}

func TestEmpresasHandler_Create(t *testing.T) {
	t.Run("responde 201 con la empresa creada", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		body := `{"nombre_empresa":"Mi Empresa"}`

		service.On("Crear", mock.Anything, mock.MatchedBy(func(p PayloadEmpresa) bool {
			return p.NombreEmpresa == "Mi Empresa" && p.userID == testUserID
		})).Return(newTestEmpresa(), nil).Once()

		r := requestWithUser(http.MethodPost, "/empresas", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyEmpresaCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el nombre esta vacio", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		body := `{"nombre_empresa":""}`

		r := requestWithUser(http.MethodPost, "/empresas", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si falta el nombre de la empresa", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		body := `{}`

		r := requestWithUser(http.MethodPost, "/empresas", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})

	t.Run("usa el ID del usuario del contexto", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		body := `{"nombre_empresa":"Mi Empresa"}`

		service.On("Crear", mock.Anything, mock.Anything).Return(model.Empresa{}, nil).Once()

		r := requestWithUser(http.MethodPost, "/empresas", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "Crear", mock.Anything, mock.MatchedBy(func(p PayloadEmpresa) bool {
			return p.userID == testUserID
		}))
	})
}

func TestEmpresasHandler_GetEmpresas(t *testing.T) {
	t.Run("responde 200 con la lista de empresas", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		empresas := []model.Empresa{newTestEmpresa()}
		service.On("ObtenerEmpresas", mock.Anything, testUserID).Return(empresas, nil).Once()

		r := requestWithUser(http.MethodGet, "/empresas", "")

		w := httptest.NewRecorder()
		handler.ObtenerEmpresas(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyEmpresasList, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando no hay empresas", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("ObtenerEmpresas", mock.Anything, testUserID).Return(nil, utilidades.ErrorNoEncontrado).Once()

		r := requestWithUser(http.MethodGet, "/empresas", "")

		w := httptest.NewRecorder()
		handler.ObtenerEmpresas(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("ObtenerEmpresas", mock.Anything, testUserID).Return(nil, assert.AnError).Once()

		r := requestWithUser(http.MethodGet, "/empresas", "")

		w := httptest.NewRecorder()
		handler.ObtenerEmpresas(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}

func TestEmpresasHandler_CambiarNombre(t *testing.T) {
	t.Run("responde 200 cuando se actualiza el nombre correctamente", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(nil).Once()

		r := requestWithNombreParam(http.MethodPatch, "/empresas/"+testEmpresaID.String()+"/Nueva Empresa", "Nueva Empresa")

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyNombreUpdated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando la empresa no existe", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithNombreParam(http.MethodPatch, "/empresas/"+testEmpresaID.String()+"/Nueva Empresa", "Nueva Empresa")

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("CambiarNombre", mock.Anything, "Nueva Empresa", testEmpresaID, testUserID).Return(assert.AnError).Once()

		r := requestWithNombreParam(http.MethodPatch, "/empresas/"+testEmpresaID.String()+"/Nueva Empresa", "Nueva Empresa")

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el nombre viene vacio", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		r := requestWithNombreParam(http.MethodPatch, "/empresas/"+testEmpresaID.String()+"/", "")

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CambiarNombre", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestEmpresasHandler_BorrarEmpresa(t *testing.T) {
	t.Run("responde 200 cuando se elimina correctamente", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(nil).Once()

		r := requestWithEmpresaContext(http.MethodDelete, "/empresas/"+testEmpresaID.String(), "")

		w := httptest.NewRecorder()
		handler.BorrarEmpresa(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyEmpresaDeleted, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando la empresa no existe", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithEmpresaContext(http.MethodDelete, "/empresas/"+testEmpresaID.String(), "")

		w := httptest.NewRecorder()
		handler.BorrarEmpresa(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(IServiceEmpresasMock)
		handler := newTestHandler(service)

		service.On("BorrarEmpresa", mock.Anything, testEmpresaID, testUserID).Return(assert.AnError).Once()

		r := requestWithEmpresaContext(http.MethodDelete, "/empresas/"+testEmpresaID.String(), "")

		w := httptest.NewRecorder()
		handler.BorrarEmpresa(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}
