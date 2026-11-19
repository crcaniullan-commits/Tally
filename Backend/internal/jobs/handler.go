package jobs

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
)

type HandlerJobs struct {
	service IServicioJobs
	errors  errores.RespuestaErrores
}

func NuevoHandlerJobs(service IServicioJobs, errors errores.RespuestaErrores) *HandlerJobs {
	return &HandlerJobs{
		service: service,
		errors:  errors,
	}
}

func (h *HandlerJobs) RecordatorioDiario(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RecordatorioDiario(r.Context()); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "recordatorio diario ejecutado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerJobs) CobrosVencidos(w http.ResponseWriter, r *http.Request) {
	if err := h.service.CobrosVencidos(r.Context()); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "aviso de cobros vencidos ejecutado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerJobs) EjecutarTodos(w http.ResponseWriter, r *http.Request) {
	if err := h.service.EjecutarTodos(r.Context()); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "todos los jobs ejecutados"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
