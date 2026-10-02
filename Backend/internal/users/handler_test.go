package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	bodyUpdateOK      = `{"data":"Usuario actualizado con exito"}`
	bodyDeleteOK      = `{"data":"Usuario eliminado"}`
	bodyNotFound      = `{"error":"not found"}`
	bodyInternalError = `{"error":"the server encountered a problem"}`
)

// newTestHandler arma un UsersHandler con un service mockeado y el logger de
// errorhandler apontando a un noop (para que los tests no ensucien la salida).
func newTestHandler(service ServiceUsers) *UsersHandler {
	return NewUserHandler(service, errorhandler.NewErrorResponse(zap.NewNop().Sugar()))
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

	return r.WithContext(context.WithValue(r.Context(), util.UserCtx, newTestUser()))
}

// requestWithRutParam inyecta el URL param "rut" que lee el handler vía
// chi.URLParam, replicando lo que hace el router real.
func requestWithRutParam(method, target, rut string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("rut", rut)

	return httptest.NewRequest(method, target, nil).WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, routeCtx),
	)
}

func decodeEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()

	var envelope map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &envelope), "respuesta no es JSON válido: %s", body)

	return envelope
}

func TestUsersHandler_Update(t *testing.T) {
	t.Run("responde 200 cuando el servicio actualiza correctamente", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		body := `{"password":"NuevaClave123","nombre":"Nombre Nuevo"}`
		usuario := newTestUser()

		service.On("Update", mock.Anything, usuario, &UpdateUserPayload{
			Password: "NuevaClave123",
			Name:     "Nombre Nuevo",
		}).Return(nil).Once()

		r := requestWithUser(http.MethodPatch, "/users", body)
		usuario = GetUserFromContext(r)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyUpdateOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 400 y no llama al servicio si la validación falla", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		// password de 3 caracteres: viola min=8
		r := requestWithUser(http.MethodPatch, "/users", `{"password":"abc","nombre":"Nombre Nuevo"}`)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Password")
		service.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("acepta una actualización parcial: solo la contraseña, sin nombre", func(t *testing.T) {
		// Nombre va con omitempty a propósito: un PATCH puede cambiar el
		// nombre, la contraseña, o las dos, pero no está obligado a mandar las
		// dos. Por eso este caso NO es 400: tiene que llegar al service, que
		// es quien decide si toca algo.
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		service.On("Update", mock.Anything, mock.Anything, &UpdateUserPayload{
			Password: "NuevaClave123",
			Name:     "",
		}).Return(nil).Once()

		r := requestWithUser(http.MethodPatch, "/users", `{"password":"NuevaClave123"}`)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyUpdateOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("acepta una actualización parcial: solo el nombre, sin contraseña", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		service.On("Update", mock.Anything, mock.Anything, &UpdateUserPayload{
			Password: "",
			Name:     "Nombre Nuevo",
		}).Return(nil).Once()

		r := requestWithUser(http.MethodPatch, "/users", `{"nombre":"Nombre Nuevo"}`)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyUpdateOK, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 y no llama al servicio si el body no es JSON válido", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPatch, "/users", `{"password":`)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 500 y no llama al servicio si viene un campo desconocido", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		// util.ReadJSON usa DisallowUnknownFields
		r := requestWithUser(http.MethodPatch, "/users", `{"password":"NuevaClave123","nombre":"Nombre","rol":"municipal"}`)

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		service.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el servicio devuelve util.ErrNotFound", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPatch, "/users", `{"password":"NuevaClave123","nombre":"Nombre Nuevo"}`)
		usuario := GetUserFromContext(r)

		service.On("Update", mock.Anything, usuario, mock.Anything).Return(util.ErrNotFound).Once()

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		r := requestWithUser(http.MethodPatch, "/users", `{"password":"NuevaClave123","nombre":"Nombre Nuevo"}`)
		usuario := GetUserFromContext(r)

		service.On("Update", mock.Anything, usuario, mock.Anything).Return(errStoreBoom).Once()

		w := httptest.NewRecorder()
		handler.Update(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_Delete(t *testing.T) {
	t.Run("responde 200 y elimina el usuario del contexto", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		usuario := newTestUser()
		service.On("Delete", mock.Anything, usuario.ID).Return(nil).Once()

		r := requestWithUser(http.MethodDelete, "/users", "")

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, bodyDeleteOK, w.Body.String())
		service.AssertExpectations(t)
		service.AssertCalled(t, "Delete", mock.Anything, usuario.ID)
	})

	t.Run("responde 404 si el usuario ya no existe", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		usuario := newTestUser()
		service.On("Delete", mock.Anything, usuario.ID).Return(util.ErrNotFound).Once()

		r := requestWithUser(http.MethodDelete, "/users", "")

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		usuario := newTestUser()
		service.On("Delete", mock.Anything, usuario.ID).Return(errStoreBoom).Once()

		r := requestWithUser(http.MethodDelete, "/users", "")

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("BUG: entra en panic si la request no trae usuario en el contexto", func(t *testing.T) {
		// handler.go:72 no valida que GetUserFromContext devuelva != nil,
		// por lo que user.ID (handler.go:74) explota con un nil pointer.
		// Este test fija el comportamiento ACTUAL para hacerlo visible.
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		r := httptest.NewRequest(http.MethodDelete, "/users", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() { handler.Delete(w, r) })
		service.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	})
}

func TestUsersHandler_GetByRut(t *testing.T) {
	t.Run("responde 200 con el usuario y el RUT parseado desde la URL", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		esperado := newTestUser()
		service.On("GetByRut", mock.Anything, util.RUT{Cuerpo: 19234567, DV: "K"}).
			Return(esperado, nil).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19.234.567-K", "19.234.567-K")

		w := httptest.NewRecorder()
		handler.GetByRut(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)

		envelope := decodeEnvelope(t, w.Body.String())
		data, ok := envelope["data"].(map[string]any)
		require.True(t, ok, "la respuesta debe traer un objeto en \"data\"")

		assert.Equal(t, esperado.ID.String(), data["id"])
		assert.Equal(t, esperado.Email, data["email"])
		assert.Equal(t, esperado.Nombre, data["nombre"])
		assert.Equal(t, float64(19234567), data["rut"].(map[string]any)["Cuerpo"])
		assert.Equal(t, "K", data["rut"].(map[string]any)["DV"])
		assert.NotContains(t, data, "PasswordHash", "la contraseña nunca debe filtrarse")
	})

	t.Run("acepta el RUT sin separadores", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		esperado := newTestUser()
		service.On("GetByRut", mock.Anything, util.RUT{Cuerpo: 19234567, DV: "K"}).
			Return(esperado, nil).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19234567K", "19234567K")

		w := httptest.NewRecorder()
		handler.GetByRut(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("responde 404 si el servicio devuelve util.ErrNotFound", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		service.On("GetByRut", mock.Anything, mock.Anything).Return(nil, util.ErrNotFound).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19.234.567-K", "19.234.567-K")

		w := httptest.NewRecorder()
		handler.GetByRut(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, bodyNotFound, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("responde 500 con cualquier otro error del servicio", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		service.On("GetByRut", mock.Anything, mock.Anything).Return(nil, errStoreBoom).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19.234.567-K", "19.234.567-K")

		w := httptest.NewRecorder()
		handler.GetByRut(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, bodyInternalError, w.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("BUG: con RUT inválido responde 400 pero igual consulta el servicio", func(t *testing.T) {
		// handler.go:98-100 escribe la respuesta 400 pero NO hace return,
		// así que sigue ejecutando la búsqueda con un RUT cero.
		// Este test fija el comportamiento ACTUAL para hacerlo visible.
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		service.On("GetByRut", mock.Anything, util.RUT{}).Return(nil, util.ErrNotFound).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/no-es-un-rut", "no-es-un-rut")

		w := httptest.NewRecorder()
		handler.GetByRut(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), util.ErrFormatoInvalido.Error())
		service.AssertCalled(t, "GetByRut", mock.Anything, util.RUT{})
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_ServiceAndHandlerWiring(t *testing.T) {
	t.Run("UserService satisface la interfaz que espera el handler", func(t *testing.T) {
		// Guarda contra regresiones de firmas: si cambia ServiceUsers, esto
		// deja de compilar.
		var _ ServiceUsers = NewUserService(new(StoreUserMock))
	})

	t.Run("el handler construye un UserService real sobre un store mockeado", func(t *testing.T) {
		store := new(StoreUserMock)
		store.On("GetByRut", mock.Anything, "19.234.567-K").Return(newTestUser(), nil).Once()

		handler := newTestHandler(NewUserService(store))

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19.234.567-K", "19.234.567-K")
		w := httptest.NewRecorder()

		handler.GetByRut(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		store.AssertExpectations(t)
	})

	t.Run("no filtra el password hash en la respuesta JSON", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		usuario := newTestUser()
		require.NoError(t, usuario.PasswordHash.Set("Secreto123"))
		service.On("GetByRut", mock.Anything, mock.Anything).Return(usuario, nil).Once()

		r := requestWithRutParam(http.MethodGet, "/users/municipal/19.234.567-K", "19.234.567-K")
		w := httptest.NewRecorder()

		handler.GetByRut(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "$2a$")
	})

	t.Run("el uuid del usuario del contexto es el que se propaga", func(t *testing.T) {
		service := new(ServiceUsersMock)
		handler := newTestHandler(service)

		usuario := newTestUser()
		service.On("Delete", mock.Anything, usuario.ID).Return(nil).Once()

		r := requestWithUser(http.MethodDelete, "/users", "")
		// sanity: el helper inyecta el usuario que esperamos
		require.Equal(t, usuario.ID, GetUserFromContext(r).ID)
		assert.NotEqual(t, uuid.Nil, usuario.ID)

		w := httptest.NewRecorder()
		handler.Delete(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		service.AssertCalled(t, "Delete", mock.Anything, uuid.MustParse(usuario.ID.String()))
	})
}
