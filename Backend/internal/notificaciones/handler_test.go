package notificaciones

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestHandler(service IServiceNotificaciones) *HandlerNotificaciones {
	return NuevoHandlerNotificaciones(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

func requestWithUserContext(method, target, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}

	usuario := &model.User{ID: testUserID, Nombre: "Emprendedor", Role: utilidades.RolUsuario}
	ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
	return r.WithContext(ctx)
}

func TestHandlerNotificaciones_ListarHistorial(t *testing.T) {
	t.Run("401 si no hay usuario en contexto", func(t *testing.T) {
		h := newTestHandler(new(IServiceNotificacionesMock))
		r := httptest.NewRequest(http.MethodGet, "/notificaciones/historial", nil)
		w := httptest.NewRecorder()

		h.ListarHistorial(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("200 con lista de historial", func(t *testing.T) {
		svc := new(IServiceNotificacionesMock)
		h := newTestHandler(svc)

		logs := []model.NotificationLog{
			{
				UserID:  testUserID,
				Kind:    "daily",
				RefDate: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC),
				SentAt:  testNow,
			},
		}

		svc.On("Listar", mock.Anything, testUserID).Return(logs, nil).Once()

		r := requestWithUserContext(http.MethodGet, "/notificaciones/historial", "")
		w := httptest.NewRecorder()

		h.ListarHistorial(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "daily")
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := new(IServiceNotificacionesMock)
		h := newTestHandler(svc)

		svc.On("Listar", mock.Anything, testUserID).Return(nil, errors.New("fallo bd")).Once()

		r := requestWithUserContext(http.MethodGet, "/notificaciones/historial", "")
		w := httptest.NewRecorder()

		h.ListarHistorial(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandlerNotificaciones_ProbarNotificacion(t *testing.T) {
	t.Run("401 si no hay usuario en contexto", func(t *testing.T) {
		h := newTestHandler(new(IServiceNotificacionesMock))
		r := httptest.NewRequest(http.MethodPost, "/notificaciones/probar", nil)
		w := httptest.NewRecorder()

		h.ProbarNotificacion(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("200 con body vacío usa valores por defecto", func(t *testing.T) {
		svc := new(IServiceNotificacionesMock)
		h := newTestHandler(svc)

		svc.On("Send", mock.Anything, testUserID, mock.MatchedBy(func(n model.Notification) bool {
			return n.Title == "Prueba de notificación"
		})).Return(nil).Once()

		r := requestWithUserContext(http.MethodPost, "/notificaciones/probar", "")
		w := httptest.NewRecorder()

		h.ProbarNotificacion(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "notificación de prueba enviada")
		svc.AssertExpectations(t)
	})

	t.Run("200 con payload personalizado", func(t *testing.T) {
		svc := new(IServiceNotificacionesMock)
		h := newTestHandler(svc)

		svc.On("Send", mock.Anything, testUserID, mock.MatchedBy(func(n model.Notification) bool {
			return n.Title == "Título personalizado" && n.Body == "Cuerpo personalizado"
		})).Return(nil).Once()

		body := `{"titulo":"Título personalizado","cuerpo":"Cuerpo personalizado"}`
		r := requestWithUserContext(http.MethodPost, "/notificaciones/probar", body)
		w := httptest.NewRecorder()

		h.ProbarNotificacion(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el envío", func(t *testing.T) {
		svc := new(IServiceNotificacionesMock)
		h := newTestHandler(svc)

		svc.On("Send", mock.Anything, testUserID, mock.Anything).
			Return(errors.New("fallo push")).Once()

		r := requestWithUserContext(http.MethodPost, "/notificaciones/probar", "")
		w := httptest.NewRecorder()

		h.ProbarNotificacion(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}
