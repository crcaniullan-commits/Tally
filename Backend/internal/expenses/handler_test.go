package expenses

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
)

const (
	bodyExpenseCreated = `{"data":{
		"id": "9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"monto": 1500,
		"category_id": "b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d",
		"descripcion": "Compra insumos",
		"fecha": "2026-09-30T00:00:00Z",
		"created_at": "2026-09-30T02:00:00Z"
	}}`
	bodyExpenseDeleted = `{"data":"expense eliminado"}`
	bodyNotFound       = `{"error":"not found"}`
)

// newTestHandler arma un ExpensesHandler con un service mockeado.
func newTestHandler(service ServiceExpenses) *ExpensesHandler {
	return NewExpensesHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
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

// requestWithExpenseIDParam inyecta el URL param "expenseID".
func requestWithExpenseIDParam(method, target, expenseID string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("expenseID", expenseID)

	r := httptest.NewRequest(method, target, nil)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, NombreEmpresa: "Mi Empresa", UserID: testUserID}
	ctx = context.WithValue(ctx, util.UserCtx, usuario)
	ctx = context.WithValue(ctx, util.EmpresaCtx, empresa)

	return r.WithContext(ctx)
}

func TestExpensesHandler_Create(t *testing.T) {
	t.Run("responde 201 con el payload creado", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"categoria":"Comida","descripcion":"Compra insumos","fecha":"2026-09-30T00:00:00Z"}`

		service.On("AddExpense", mock.Anything, mock.MatchedBy(func(p ExpensePayload) bool {
			return p.Monto == 1500 && p.Categoria != nil && *p.Categoria == testCategoryName
		}), testEmpresaID).Return(newTestExpense(), nil).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyExpenseCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("usa el ID de la empresa del contexto", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`

		service.On("AddExpense", mock.Anything, mock.Anything, testEmpresaID).Return(model.Expense{}, nil).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "AddExpense", mock.Anything, mock.Anything, testEmpresaID)
	})

	t.Run("responde 400 y no llama al servicio si el monto es cero", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", `{"monto":0,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el monto es negativo", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", `{"monto":-50,"category_id":"b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"}`)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "AddExpense", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si la categoria no existe", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"categoria":"Inexistente"}`

		service.On("AddExpense", mock.Anything, mock.Anything, testEmpresaID).Return(model.Expense{}, util.ErrNotFound).Once()

		r := requestWithEmpresaContext(http.MethodPost, "/expenses", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertExpectations(t)
	})
}

func TestExpensesHandler_Delete(t *testing.T) {
	t.Run("responde 200 cuando se elimina correctamente", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("DeleteExpense", mock.Anything, testExpenseID, testEmpresaID).Return(nil).Once()

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/"+testExpenseID.String(), testExpenseID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyExpenseDeleted, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 404 cuando el gasto no existe", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		service.On("DeleteExpense", mock.Anything, testExpenseID, testEmpresaID).Return(util.ErrNotFound).Once()

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/"+testExpenseID.String(), testExpenseID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el expenseID es invalido", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		r := requestWithExpenseIDParam(http.MethodDelete, "/expenses/invalid", "invalid")

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "DeleteExpense", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestExpensesHandler_GetExpensesOfEmpresa(t *testing.T) {
	t.Run("responde 200 con la lista de gastos", func(t *testing.T) {
		service := new(ServiceExpensesMock)
		handler := newTestHandler(service)

		expenses := []model.Expense{newTestExpense()}
		service.On("GetAllExpensesOfEmpresa", mock.Anything, testEmpresaID, mock.AnythingOfType("pagination.IncomePaginationQuery")).
			Return(expenses, nil).Once()

		r := requestWithEmpresaContext(http.MethodGet, "/expenses", "")

		w := httptest.NewRecorder()
		handler.GetExpensesOfEmpresa(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
	})
}
