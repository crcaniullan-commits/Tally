package expenses

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
	bodyExpenseCreated = `{"datos":{
		"id": "9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"monto": 1500,
		"category_id": "b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d",
		"descripcion": "Compra insumos",
		"fecha": "2026-09-30T00:00:00Z",
		"created_at": "2026-09-30T02:00:00Z"
	}}`
	bodyExpenseDeleted = `{"datos":"gasto eliminado"}`
	bodyNotFound       = `{"error":"no encontrado"}`
)

// newTestHandler arma un HandlerGastos con un service mockeado.
func newTestHandler(service IServiceGastos) *HandlerGastos {
	return NuevoHandlerGastos(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
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

// requestWithExpenseIDParam inyecta el URL param "gastoID".
func requestWithExpenseIDParam(method, target, gastoID string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("gastoID", gastoID)

	r := httptest.NewRequest(method, target, nil)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
	ctx = context.WithValue(ctx, utilidades.ContextoUsuario, usuario)
	ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

	return r.WithContext(ctx)
}

func TestExpensesHandler_Create(t *testing.T) {
	t.Run("responde 201 con el payload creado", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d","descripcion":"Compra insumos"}`

		service.On("AgregarGasto", mock.Anything, mock.MatchedBy(func(p PayloadGasto) bool {
			return p.Monto == 1500 && p.CategoryID != nil && *p.CategoryID == testCategoryID
		}), testEmpresaID).Return(newTestExpense(), nil).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyExpenseCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("usa el ID de la empresa del contexto", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`

		service.On("AgregarGasto", mock.Anything, mock.Anything, testEmpresaID).Return(model.Expense{}, nil).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "AgregarGasto", mock.Anything, mock.Anything, testEmpresaID)
	})

	t.Run("responde 400 y no llama al servicio si el monto es cero", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", `{"monto":0,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el monto es negativo", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", `{"monto":-50,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si la categoria no existe", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`

		service.On("AgregarGasto", mock.Anything, mock.Anything, testEmpresaID).Return(model.Expense{}, utilidades.ErrorNoEncontrado).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si falta category_id", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", `{"monto":1500}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "CategoryID")
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si category_id no es un UUID", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses",
			`{"monto":1500,"category_id":"Comida"}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si mandan fecha en el payload", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses",
			`{"monto":1500,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d","fecha":"2026-09-30T00:00:00Z"}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "fecha")
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si mandan la categoria por nombre", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses",
			`{"monto":1500,"categoria":"Comida"}`)

		w := httptest.NewRecorder()
		handler.Crear(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "categoria")
		service.AssertNotCalled(t, "AgregarGasto", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestExpensesHandler_Delete(t *testing.T) {
	t.Run("responde 200 cuando se elimina correctamente", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("EliminarGasto", mock.Anything, testExpenseID, testEmpresaID).Return(nil).Once()

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/"+testExpenseID.String(), testExpenseID.String())

		w := httptest.NewRecorder()
		handler.Eliminar(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyExpenseDeleted, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando el gasto no existe", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("EliminarGasto", mock.Anything, testExpenseID, testEmpresaID).Return(utilidades.ErrorNoEncontrado).Once()

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/"+testExpenseID.String(), testExpenseID.String())

		w := httptest.NewRecorder()
		handler.Eliminar(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el gastoID es invalido", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/invalid", "invalid")

		w := httptest.NewRecorder()
		handler.Eliminar(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "EliminarGasto", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestExpensesHandler_GetExpensesOfEmpresa(t *testing.T) {
	t.Run("responde 200 con la lista de gastos", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		expenses := []model.Expense{newTestExpense()}
		service.On("ObtenerGastosDeEmpresa", mock.Anything, testEmpresaID, mock.AnythingOfType("paginacion.ConsultaPaginacionIngresos")).
			Return(expenses, nil).Once()

		r := requestWithEmpresaContext(http.MethodGet, "/expenses", "")

		w := httptest.NewRecorder()
		handler.ObtenerGastosDeEmpresa(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
	})
}

func TestExpensesHandler_Update(t *testing.T) {
	// updateRequest arma una request PATCH con el URL param "gastoID" y el
	// contexto del middleware de empresa.
	updateRequest := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPatch, "/gastos/"+testExpenseID.String(), strings.NewReader(body))

		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("gastoID", testExpenseID.String())

		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
		usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
		empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
		ctx = context.WithValue(ctx, utilidades.ContextoUsuario, usuario)
		ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

		return req.WithContext(ctx)
	}

	t.Run("responde 200 con el gasto actualizado", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		monto := int64(2500)
		fecha := "2026-09-15"

		service.On("ActualizarGasto", mock.Anything, PayloadActualizarGasto{
			Monto: &monto,
			Fecha: &fecha,
		}, testExpenseID, testEmpresaID).Return(newTestExpense(), nil).Once()

		w := httptest.NewRecorder()
		handler.Actualizar(w, updateRequest(`{"monto":2500,"fecha":"2026-09-15"}`))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"monto":1500`)
		service.AssertExpectations(t)
	})

	t.Run("responde 404 si el gasto no existe", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.Expense{}, ErrorMovimientoNoEncontrado).Once()

		w := httptest.NewRecorder()
		handler.Actualizar(w, updateRequest(`{"monto":2500}`))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
	})

	t.Run("responde 400 con una fecha invalida", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.Expense{}, ErrorFechaInvalida).Once()

		w := httptest.NewRecorder()
		handler.Actualizar(w, updateRequest(`{"fecha":"15-09-2026"}`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"la fecha debe tener el formato AAAA-MM-DD"}`, w.Body.String())
	})

	t.Run("responde 400 si el gastoID no es un uuid", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		req := httptest.NewRequest(http.MethodPatch, "/gastos/no-es-uuid", strings.NewReader(`{"monto":2500}`))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("gastoID", "no-es-uuid")
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)

		w := httptest.NewRecorder()
		handler.Actualizar(w, req.WithContext(ctx))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el payload no tiene campos", func(t *testing.T) {
		service := new(IServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("ActualizarGasto", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.Expense{}, ErrorSinCambios).Once()

		w := httptest.NewRecorder()
		handler.Actualizar(w, updateRequest(`{}`))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"no se enviaron campos para actualizar"}`, w.Body.String())
	})
}
