package incomes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
)

const (
	bodyIncomeCreated = `{"data":{
		"id": "9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e",
		"userID": "6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b",
		"monto": 1500,
		"payment_method": "debito",
		"descripcion": "Venta de almuerzo",
		"fecha": "2026-09-30T00:00:00Z",
		"created_at": "2026-09-30T02:00:00Z"
	}}`
	bodyIncomeDeleted = `{"data":"income eliminado"}`
	bodyNotFound      = `{"error":"not found"}`
	bodyInternalError = `{"error":"the server encountered a problem"}`
)

// newTestHandler arma un IncomesHandler con un service mockeado y el logger de
// errorhandler apontando a un noop (para que los tests no ensucien la salida).
func newTestHandler(service ServiceIncomes) *IncomesHandler {
	return NewIncomesHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
}

// requestWithUser construye una request con el usuario del middleware inyectado
// en el contexto, igual que lo hace el middleware de autenticación.
func requestWithUser(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &users.Users{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}

	return r.WithContext(context.WithValue(r.Context(), util.UserCtx, usuario))
}

// requestWithIncomeIDParam inyecta el URL param "incomeID" que lee el handler
// vía chi.URLParam, replicando lo que hace el router real.
func requestWithIncomeIDParam(method, target, incomeID string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("incomeID", incomeID)

	r := httptest.NewRequest(method, target, nil).WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, routeCtx),
	)

	usuario := &users.Users{ID: testUserID, Nombre: "Emprendedor", Role: util.UserRoleUsuario}

	return r.WithContext(context.WithValue(r.Context(), util.UserCtx, usuario))
}

func decodeEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()

	var envelope map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &envelope), "respuesta no es JSON válido: %s", body)

	return envelope
}

func TestIncomesHandler_AddIncome(t *testing.T) {
	t.Run("responde 201 con el payload creado", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"payment_method":"debito","descripcion":"Venta de almuerzo"}`

		service.On("AddIncome", mock.Anything, IncomePayload{
			Monto:         1500,
			PaymentMethod: util.PaymentMethodDebito,
			Descripcion:   ptr("Venta de almuerzo"),
		}, testUserID).Return(newTestIncome(), nil).Once()

		r := requestWithUser(http.MethodPost, "/incomes", body)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, bodyIncomeCreated, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("usa el ID del usuario del contexto, no uno del body", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		body := `{"monto":1500,"payment_method":"debito"}`

		service.On("AddIncome", mock.Anything, mock.Anything, testUserID).Return(IncomeStorage{}, nil).Once()

		r := requestWithUser(http.MethodPost, "/incomes", body)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		service.AssertCalled(t, "AddIncome", mock.Anything, mock.Anything, testUserID)
	})

	t.Run("acepta los cuatro metodos de pago", func(t *testing.T) {
		metodos := []string{"debito", "credito", "transferencia", "efectivo"}

		for _, metodo := range metodos {
			t.Run(metodo, func(t *testing.T) {
				service := new(ServiceIncomesMock)
				handler := newTestHandler(service)

				body := `{"monto":100,"payment_method":"` + metodo + `"}`
				service.On("AddIncome", mock.Anything, mock.Anything, mock.Anything).
					Return(IncomeStorage{}, nil).Once()

				r := requestWithUser(http.MethodPost, "/incomes", body)

				w := httptest.NewRecorder()
				handler.AddIncome(w, r)

				assert.Equal(t, http.StatusCreated, w.Code)
				service.AssertCalled(t, "AddIncome", mock.Anything,
					mock.MatchedBy(func(p IncomePayload) bool {
						return string(p.PaymentMethod) == metodo
					}), testUserID)
			})
		}
	})

	t.Run("responde 400 y no llama al servicio si el monto es cero", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":0,"payment_method":"debito"}`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Monto")
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el monto es negativo", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":-50,"payment_method":"debito"}`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "gt")
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si falta el metodo de pago", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":100}`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "PaymentMethod")
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el metodo de pago no es valido", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":100,"payment_method":"bitcoin"}`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "oneof")
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si la descripcion supera los 100 caracteres", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		larga := strings.Repeat("a", 101)
		body, err := json.Marshal(map[string]any{
			"monto":          100,
			"payment_method": "debito",
			"descripcion":    larga,
		})
		require.NoError(t, err)

		r := requestWithUser(http.MethodPost, "/incomes", string(body))

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Descripcion")
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 y no llama al servicio si el body no es JSON valido", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 y no llama al servicio si viene un campo desconocido", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		// util.ReadJSON usa DisallowUnknownFields
		r := requestWithUser(http.MethodPost, "/incomes",
			`{"monto":100,"payment_method":"debito","user_id":"6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b"}`)

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 si el servicio falla", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":100,"payment_method":"debito"}`)
		service.On("AddIncome", mock.Anything, mock.Anything, mock.Anything).
			Return(IncomeStorage{}, errStoreBoom).Once()

		w := httptest.NewRecorder()
		handler.AddIncome(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestIncomesHandler_Delete(t *testing.T) {
	t.Run("responde 200 y elimina el ingreso del URL param", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(nil).Once()

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/"+testIncomeID.String(), testIncomeID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyIncomeDeleted, w.Body.String())
		service.AssertExpectations(t)
		service.AssertCalled(t, "DeleteIncome", mock.Anything, testIncomeID, testUserID)
	})

	t.Run("responde 400 y no llama al servicio si el incomeID no es un UUID", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/no-es-uuid", "no-es-uuid")

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "DeleteIncome", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el ingreso no existe", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(util.ErrNotFound).Once()

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/"+testIncomeID.String(), testIncomeID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).Return(errStoreBoom).Once()

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/"+testIncomeID.String(), testIncomeID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("pasa el userID del contexto para que el store acote el borrado", func(t *testing.T) {
		// El handler tiene que adjuntar el userID del token. Si solo mandara el
		// incomeID de la URL, el store no podria distinguir de quien es el
		// ingreso y responderia 200 al borrar algo ajeno.
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		victima := newTestIncome()
		service.On("DeleteIncome", mock.Anything, victima.ID, testUserID).Return(nil).Once()

		r := requestWithIncomeIDParam(http.MethodDelete,
			"/incomes/"+victima.ID.String(), victima.ID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "DeleteIncome", mock.Anything, victima.ID, testUserID)
	})

	t.Run("responde 404 cuando el ingreso es de otro usuario", func(t *testing.T) {
		// El store acotó por user_id, no matcheó filas y devolvió ErrNotFound.
		// El handler lo traduce a 404 en vez de 200.
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("DeleteIncome", mock.Anything, testIncomeID, testUserID).
			Return(util.ErrNotFound).Once()

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/"+testIncomeID.String(), testIncomeID.String())

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestIncomesHandler_GetIncomesOfUser(t *testing.T) {
	t.Run("responde 200 con la lista de ingresos del usuario del contexto", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		esperados := []IncomeStorage{newTestIncome()}
		service.On("GetAllIncomesOfUser", mock.Anything, testUserID, newTestFilterQuery()).
			Return(esperados, nil).Once()

		r := requestWithUser(http.MethodGet, "/incomes", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)

		envelope := decodeEnvelope(t, w.Body.String())
		data, ok := envelope["data"].([]any)
		require.True(t, ok, "la respuesta debe traer un arreglo en \"data\"")

		require.Len(t, data, 1)
		ingreso := data[0].(map[string]any)

		assert.Equal(t, testIncomeID.String(), ingreso["id"])
		assert.Equal(t, testUserID.String(), ingreso["userID"])
		assert.Equal(t, float64(1500), ingreso["monto"])
		assert.Equal(t, "debito", ingreso["payment_method"])
		assert.Equal(t, "Venta de almuerzo", ingreso["descripcion"])
	})

	t.Run("sin query params pide la primera página del tamaño tope", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("GetAllIncomesOfUser", mock.Anything, testUserID,
			mock.MatchedBy(func(fq pagination.IncomePaginationQuery) bool {
				return fq.Limit == pagination.DefaultLimit &&
					fq.Offset == 0 &&
					fq.Since == "" &&
					fq.Until == ""
			}),
		).Return([]IncomeStorage{}, nil).Once()

		r := requestWithUser(http.MethodGet, "/incomes", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("manda al servicio los query params de paginación y rango", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		esperada := pagination.IncomePaginationQuery{
			Limit:  5,
			Offset: 10,
			Since:  "2026-09-01",
			Until:  "2026-09-30",
		}

		service.On("GetAllIncomesOfUser", mock.Anything, testUserID, esperada).
			Return([]IncomeStorage{}, nil).Once()

		r := requestWithUser(http.MethodGet,
			"/incomes?limit=5&offset=10&since=2026-09-01&until=2026-09-30", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
		service.AssertCalled(t, "GetAllIncomesOfUser", mock.Anything, testUserID, esperada)
	})

	t.Run("responde 400 y no llama al servicio si limit no es un entero", func(t *testing.T) {
		// Si el Parse se comiera el error, el cliente pediría una página y
		// recibiría otra sin enterarse.
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodGet, "/incomes?limit=abc", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "limit")
		service.AssertNotCalled(t, "GetAllIncomesOfUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si limit se pasa del máximo", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodGet, "/incomes?limit=500", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "GetAllIncomesOfUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si offset es negativo", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodGet, "/incomes?offset=-1", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		service.AssertNotCalled(t, "GetAllIncomesOfUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si la fecha del rango no es AAAA-MM-DD", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodGet, "/incomes?since=30-09-2026", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "since")
		service.AssertNotCalled(t, "GetAllIncomesOfUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 200 con data null cuando el usuario no tiene ingresos", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("GetAllIncomesOfUser", mock.Anything, testUserID, mock.Anything).
			Return(nil, nil).Once()

		r := requestWithUser(http.MethodGet, "/incomes", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		envelope := decodeEnvelope(t, w.Body.String())

		assert.Nil(t, envelope["data"], "slice nil se serializa como null")
		service.AssertExpectations(t)
	})

	t.Run("responde 200 con data vacio cuando el store devuelve slice vacío", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("GetAllIncomesOfUser", mock.Anything, testUserID, mock.Anything).
			Return([]IncomeStorage{}, nil).Once()

		r := requestWithUser(http.MethodGet, "/incomes", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		envelope := decodeEnvelope(t, w.Body.String())

		assert.Equal(t, []any{}, envelope["data"])
		service.AssertExpectations(t)
	})

	t.Run("responde 500 si el servicio falla", func(t *testing.T) {
		service := new(ServiceIncomesMock)
		handler := newTestHandler(service)

		service.On("GetAllIncomesOfUser", mock.Anything, testUserID, mock.Anything).
			Return(nil, errStoreBoom).Once()

		r := requestWithUser(http.MethodGet, "/incomes", "")

		w := httptest.NewRecorder()
		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestIncomesHandler_ServiceAndHandlerWiring(t *testing.T) {
	t.Run("IncomeService satisface la interfaz que espera el handler", func(t *testing.T) {
		// Guarda contra regresiones de firmas: si cambia ServiceIncomes, esto
		// deja de compilar.
		var _ ServiceIncomes = NewIncomeService(new(StoreIncomesMock))
	})

	t.Run("el handler construye un IncomeService real sobre un store mockeado", func(t *testing.T) {
		store := new(StoreIncomesMock)
		store.On("GetAllIncomesOfUser", mock.Anything, testUserID, newTestFilterQuery()).
			Return([]IncomeStorage{newTestIncome()}, nil).Once()

		handler := newTestHandler(NewIncomeService(store))

		r := requestWithUser(http.MethodGet, "/incomes", "")
		w := httptest.NewRecorder()

		handler.GetIncomesOfUser(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		store.AssertExpectations(t)
	})

	t.Run("el alta llega al store con el userID del contexto", func(t *testing.T) {
		store := new(StoreIncomesMock)
		store.On("AddIncome", mock.Anything, mock.Anything).Return(nil).Once()

		handler := newTestHandler(NewIncomeService(store))

		r := requestWithUser(http.MethodPost, "/incomes", `{"monto":100,"payment_method":"efectivo"}`)
		w := httptest.NewRecorder()

		handler.AddIncome(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		store.AssertCalled(t, "AddIncome", mock.Anything,
			mock.MatchedBy(func(i *IncomeStorage) bool {
				return i.UserID == testUserID &&
					i.Monto == 100 &&
					i.PaymentMethod == util.PaymentMethodEfectivo &&
					i.Descripcion == nil
			}))
	})

	t.Run("el UUID del income propagado coincide con el de la URL", func(t *testing.T) {
		store := new(StoreIncomesMock)
		otroID := uuid.New()
		store.On("DeleteIncome", mock.Anything, otroID, testUserID).Return(nil).Once()

		handler := newTestHandler(NewIncomeService(store))

		r := requestWithIncomeIDParam(http.MethodDelete, "/incomes/"+otroID.String(), otroID.String())
		w := httptest.NewRecorder()

		handler.Delete(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		store.AssertCalled(t, "DeleteIncome", mock.Anything, otroID, testUserID)
	})
}

func ptr(s string) *string { return &s }
