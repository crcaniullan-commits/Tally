package jobs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestHandler(service IServicioJobs) *HandlerJobs {
	return NuevoHandlerJobs(service, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

func TestHandlerJobs_RecordatorioDiario(t *testing.T) {
	t.Run("200 al ejecutar exitosamente", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("RecordatorioDiario", mock.Anything).Return(nil).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/recordatorio-diario", nil)
		w := httptest.NewRecorder()

		h.RecordatorioDiario(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "recordatorio diario ejecutado")
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("RecordatorioDiario", mock.Anything).Return(errors.New("fallo")).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/recordatorio-diario", nil)
		w := httptest.NewRecorder()

		h.RecordatorioDiario(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandlerJobs_CobrosVencidos(t *testing.T) {
	t.Run("200 al ejecutar exitosamente", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("CobrosVencidos", mock.Anything).Return(nil).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/cobros-vencidos", nil)
		w := httptest.NewRecorder()

		h.CobrosVencidos(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "aviso de cobros vencidos ejecutado")
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("CobrosVencidos", mock.Anything).Return(errors.New("fallo")).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/cobros-vencidos", nil)
		w := httptest.NewRecorder()

		h.CobrosVencidos(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandlerJobs_EjecutarTodos(t *testing.T) {
	t.Run("200 al ejecutar exitosamente", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("EjecutarTodos", mock.Anything).Return(nil).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/ejecutar-todos", nil)
		w := httptest.NewRecorder()

		h.EjecutarTodos(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "todos los jobs ejecutados")
		svc.AssertExpectations(t)
	})

	t.Run("500 si falla el servicio", func(t *testing.T) {
		svc := new(IServicioJobsMock)
		h := newTestHandler(svc)

		svc.On("EjecutarTodos", mock.Anything).Return(errors.New("fallo")).Once()

		r := httptest.NewRequest(http.MethodPost, "/jobs/ejecutar-todos", nil)
		w := httptest.NewRecorder()

		h.EjecutarTodos(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}
