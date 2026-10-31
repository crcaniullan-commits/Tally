package categories

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
)

const (
	bodyCategoriaCreated = `{"data":{
		"id": "b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre": "Comida",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
	}}`
	bodyNombreCambiado     = `{"data":"Nombre cambiado"}`
	bodyCategoriaEliminada = `{"data":"Categoria b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d eliminada"}`
	bodyNotFound           = `{"error":"not found"}`
	bodyBadRequest         = `{"error":"La categoría ya existe"}`
)

// newTestHandler arma un CategoriesHandler con un service mockeado y el logger de
// errorhandler apuntando a un noop.
func newTestHandler(service ServiceCategories) *CategoriesHandler {
	return NewCategoriesHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
}

// requestWithEmpresaContext inyecta usuario y empresa en el contexto.
func requestWithEmpresaContext(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}

	ctx := context.WithValue(r.Context(), util.UserCtx, usuario)
	ctx = context.WithValue(ctx, util.EmpresaCtx, empresa)

	return r.WithContext(ctx)
}

// requestWithCategoriaIDParam inyecta el URL param "categoriaID" que lee el handler.
func requestWithCategoriaIDParam(method, target, categoriaID string, body string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("categoriaID", categoriaID)

	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
	ctx = context.WithValue(ctx, util.UserCtx, usuario)
	ctx = context.WithValue(ctx, util.EmpresaCtx, empresa)

	return r.WithContext(ctx)
}

// requestWithCategoriaParam inyecta el URL param "categoria" que lee el handler.
func requestWithCategoriaParam(method, target, categoria string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("categoria", categoria)

	r := httptest.NewRequest(method, target, nil)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
	ctx = context.WithValue(ctx, util.UserCtx, usuario)
	ctx = context.WithValue(ctx, util.EmpresaCtx, empresa)

	return r.WithContext(ctx)
}

func TestCategoriesHandler_Create(t *testing.T) {
	t.Run("responde 201 con la categoria creada", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Comida"}`
		empresaID := testEmpresaID

		service.On("Create", mock.Anything, CategoriaPayload{Nombre: "Comida"}, &empresaID).
			Return(newTestCategoria(), nil).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/categories", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyCategoriaCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si la categoria ya existe", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Comida"}`
		empresaID := testEmpresaID

		service.On("Create", mock.Anything, CategoriaPayload{Nombre: "Comida"}, &empresaID).
			Return(model.Categorie{}, AlreadyExistError).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/categories", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, bodyBadRequest, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el nombre esta vacio", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":""}`

		r := requestWithEmpresaContext(http.MethodPost, "/categories", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si falta el nombre", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{}`

		r := requestWithEmpresaContext(http.MethodPost, "/categories", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Comida"}`
		empresaID := testEmpresaID

		service.On("Create", mock.Anything, CategoriaPayload{Nombre: "Comida"}, &empresaID).
			Return(model.Categorie{}, assert.AnError).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/categories", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}

func TestCategoriesHandler_GetCategorias(t *testing.T) {
	t.Run("responde 200 con las categorias de la empresa y las globales", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		global := model.Categorie{
			ID:     uuid.MustParse("c1d2e3f4-5a6b-4c7d-8e9f-0a1b2c3d4e5f"),
			Nombre: "Ventas",
		}

		service.On("GetCategorias", mock.Anything, testEmpresaID).
			Return([]model.Categorie{newTestCategoria(), global}, nil).Once()

		r := requestWithEmpresaContext(http.MethodGet, "/categories", "")

		w := httptest.NewRecorder()
		handler.GetCategorias(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"data":[
			{"id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d","nombre":"Comida","empresa_id":"1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"},
			{"id":"c1d2e3f4-5a6b-4c7d-8e9f-0a1b2c3d4e5f","nombre":"Ventas","empresa_id":null}
		]}`, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		service.On("GetCategorias", mock.Anything, testEmpresaID).
			Return(nil, assert.AnError).Once()

		r := requestWithEmpresaContext(http.MethodGet, "/categories", "")

		w := httptest.NewRecorder()
		handler.GetCategorias(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}

func TestCategoriesHandler_CambiarNombre(t *testing.T) {
	t.Run("responde 200 cuando se cambia el nombre correctamente", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Bebidas"}`

		service.On("CambiarNombre", mock.Anything, CategoriaPayload{Nombre: "Bebidas"}, testCategoriaID, testEmpresaID).
			Return(nil).Once()

		r := requestWithCategoriaIDParam(http.MethodPatch, "/categories/"+testCategoriaID.String(), testCategoriaID.String(), body)

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyNombreCambiado, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando la categoria no existe", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Bebidas"}`

		service.On("CambiarNombre", mock.Anything, CategoriaPayload{Nombre: "Bebidas"}, testCategoriaID, testEmpresaID).
			Return(util.ErrNotFound).Once()

		r := requestWithCategoriaIDParam(http.MethodPatch, "/categories/"+testCategoriaID.String(), testCategoriaID.String(), body)

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el nombre ya existe en la empresa", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Bebidas"}`

		service.On("CambiarNombre", mock.Anything, CategoriaPayload{Nombre: "Bebidas"}, testCategoriaID, testEmpresaID).
			Return(AlreadyExistError).Once()

		r := requestWithCategoriaIDParam(http.MethodPatch, "/categories/"+testCategoriaID.String(), testCategoriaID.String(), body)

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, bodyBadRequest, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el categoriaID es invalido", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Bebidas"}`

		r := requestWithCategoriaIDParam(http.MethodPatch, "/categories/invalid", "invalid", body)

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "CambiarNombre", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Bebidas"}`

		service.On("CambiarNombre", mock.Anything, CategoriaPayload{Nombre: "Bebidas"}, testCategoriaID, testEmpresaID).
			Return(assert.AnError).Once()

		r := requestWithCategoriaIDParam(http.MethodPatch, "/categories/"+testCategoriaID.String(), testCategoriaID.String(), body)

		w := httptest.NewRecorder()
		handler.CambiarNombre(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}

func TestCategoriesHandler_BorrarCategoria(t *testing.T) {
	t.Run("responde 200 cuando se elimina correctamente", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		service.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(nil).Once()

		r := requestWithCategoriaParam(http.MethodDelete, "/categories/"+testCategoriaID.String(), testCategoriaID.String())

		w := httptest.NewRecorder()
		handler.BorrarCategoria(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyCategoriaEliminada, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando la categoria no existe", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		service.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(util.ErrNotFound).Once()

		r := requestWithCategoriaParam(http.MethodDelete, "/categories/"+testCategoriaID.String(), testCategoriaID.String())

		w := httptest.NewRecorder()
		handler.BorrarCategoria(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el categoriaID es invalido", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		r := requestWithCategoriaParam(http.MethodDelete, "/categories/invalid", "invalid")

		w := httptest.NewRecorder()
		handler.BorrarCategoria(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "BorrarCategoria", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 en caso de error interno", func(t *testing.T) {
		service := new(ServiceCategoriesMock)
		handler := newTestHandler(service)

		service.On("BorrarCategoria", mock.Anything, testCategoriaID, testEmpresaID).Return(assert.AnError).Once()

		r := requestWithCategoriaParam(http.MethodDelete, "/categories/"+testCategoriaID.String(), testCategoriaID.String())

		w := httptest.NewRecorder()
		handler.BorrarCategoria(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertExpectations(t)
	})
}
