package debtors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const (
	bodyDebtorCreated = `{"data":{
		"id": "f6a7b8c9-0d1e-4f2a-8b3c-4d5e6f7a8b9c",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre": "Cliente",
		"monto": 15000,
		"descripcion": "Deuda pendiente",
		"pagado": false,
		"fecha_limite": "2026-11-30T00:00:00Z",
		"created_at": "2026-11-05T12:00:00Z"
	}}`
	bodyDebtorList = `{"data":[{
		"id": "f6a7b8c9-0d1e-4f2a-8b3c-4d5e6f7a8b9c",
		"empresa_id": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		"nombre": "Cliente",
		"monto": 15000,
		"descripcion": "Deuda pendiente",
		"pagado": false,
		"fecha_limite": "2026-11-30T00:00:00Z",
		"created_at": "2026-11-05T12:00:00Z"
	}]}`
	bodyDebtorUpdated = `{"data":"Deuda actualizada"}`
	bodyDebtorPaid    = `{"data":"Deuda pagada"}`
	bodyNotFound      = `{"error":"not found"}`
	bodyInternalError = `{"error":"the server encountered a problem"}`
)

func newTestHandler(service ServiceDebtors) *DebtorsHandler {
	return NewDebtorsHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
}

func newTestEmpresaContext(r *http.Request) *http.Request {
	usuario := &model.User{ID: testUserID, Role: util.UserRoleUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, UserID: testUserID}

	ctx := context.WithValue(r.Context(), util.UserCtx, usuario)
	ctx = context.WithValue(ctx, util.EmpresaCtx, empresa)

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

func requestWithDebtorID(method, target, debtorID, body string) *http.Request {
	r := requestWithEmpresa(method, target, body)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("debtorID", debtorID)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)

	return r.WithContext(ctx)
}

func TestDebtorsHandler_Create(t *testing.T) {
	t.Run("responde 201 con la deuda creada", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		payload := newTestDebtorPayload()

		service.On("Create", mock.Anything, testEmpresaID, payload).
			Return(newTestDebtor(), nil).Once()

		body := `{"nombre":"Cliente","monto":15000,"descripcion":"Deuda pendiente","fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithEmpresa(http.MethodPost, "/debtors", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyDebtorCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el JSON es invalido", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		r := requestWithEmpresa(http.MethodPost, "/debtors", `{"nombre":`)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el monto no es mayor que cero", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Cliente","monto":0,"descripcion":"Deuda pendiente","fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithEmpresa(http.MethodPost, "/debtors", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si falta un campo requerido", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Cliente","monto":15000,"descripcion":"Deuda pendiente"}`
		r := requestWithEmpresa(http.MethodPost, "/debtors", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si hay un campo desconocido", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Cliente","monto":15000,"descripcion":"Deuda pendiente","fecha_limite":"2026-11-30T00:00:00Z","extra":1}`
		r := requestWithEmpresa(http.MethodPost, "/debtors", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("Create", mock.Anything, testEmpresaID, mock.Anything).
			Return(model.Debtor{}, errStoreBoom).Once()

		body := `{"nombre":"Cliente","monto":15000,"descripcion":"Deuda pendiente","fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithEmpresa(http.MethodPost, "/debtors", body)

		w := httptest.NewRecorder()
		handler.Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestDebtorsHandler_Update(t *testing.T) {
	t.Run("responde 200 con el mensaje de confirmacion", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		payload := newTestUpdatePayload()

		service.On("Update", mock.Anything, testDebtorID, testEmpresaID, payload).
			Return(nil).Once()

		body := `{"nombre":"Cliente","monto":15000,"fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String(), testDebtorID.String(), body)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyDebtorUpdated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el debtorID no es un UUID", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Cliente","monto":15000,"fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithDebtorID(http.MethodPatch, "/debtors/no-es-uuid", "no-es-uuid", body)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el payload no pasa la validacion", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		body := `{"nombre":"Cliente","monto":0,"fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String(), testDebtorID.String(), body)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si la deuda no existe", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("Update", mock.Anything, testDebtorID, testEmpresaID, mock.Anything).
			Return(util.ErrNotFound).Once()

		body := `{"nombre":"Cliente","monto":15000,"fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String(), testDebtorID.String(), body)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("Update", mock.Anything, testDebtorID, testEmpresaID, mock.Anything).
			Return(errStoreBoom).Once()

		body := `{"nombre":"Cliente","monto":15000,"fecha_limite":"2026-11-30T00:00:00Z"}`
		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String(), testDebtorID.String(), body)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestDebtorsHandler_PaidDebtor(t *testing.T) {
	t.Run("responde 200 con el mensaje de confirmacion", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("PaidDebtor", mock.Anything, testDebtorID, testEmpresaID).Return(nil).Once()

		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String()+"/paid", testDebtorID.String(), "")

		w := httptest.NewRecorder()
		handler.PaidDebtor(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyDebtorPaid, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 si el debtorID no es un UUID", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		r := requestWithDebtorID(http.MethodPatch, "/debtors/no-es-uuid/paid", "no-es-uuid", "")

		w := httptest.NewRecorder()
		handler.PaidDebtor(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "PaidDebtor", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si la deuda no existe", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("PaidDebtor", mock.Anything, testDebtorID, testEmpresaID).
			Return(util.ErrNotFound).Once()

		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String()+"/paid", testDebtorID.String(), "")

		w := httptest.NewRecorder()
		handler.PaidDebtor(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("PaidDebtor", mock.Anything, testDebtorID, testEmpresaID).
			Return(errStoreBoom).Once()

		r := requestWithDebtorID(http.MethodPatch, "/debtors/"+testDebtorID.String()+"/paid", testDebtorID.String(), "")

		w := httptest.NewRecorder()
		handler.PaidDebtor(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestDebtorsHandler_GetDebtorsOfEmpresa(t *testing.T) {
	t.Run("responde 200 con la lista de deudas", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("GetDebtors", mock.Anything, testEmpresaID).
			Return([]model.Debtor{newTestDebtor()}, nil).Once()

		r := requestWithEmpresa(http.MethodGet, "/debtors", "")

		w := httptest.NewRecorder()
		handler.GetDebtorsOfEmpresa(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyDebtorList, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 200 con lista vacia", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("GetDebtors", mock.Anything, testEmpresaID).
			Return([]model.Debtor{}, nil).Once()

		r := requestWithEmpresa(http.MethodGet, "/debtors", "")

		w := httptest.NewRecorder()
		handler.GetDebtorsOfEmpresa(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"data":[]}`, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		service := new(ServiceDebtorsMock)
		handler := newTestHandler(service)

		service.On("GetDebtors", mock.Anything, testEmpresaID).
			Return(nil, errStoreBoom).Once()

		r := requestWithEmpresa(http.MethodGet, "/debtors", "")

		w := httptest.NewRecorder()
		handler.GetDebtorsOfEmpresa(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}
