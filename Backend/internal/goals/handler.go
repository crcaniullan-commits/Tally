package goals

import (
	"context"
	"errors"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceMetas interface {
	CrearMeta(context.Context, PayloadCrearMeta, uuid.UUID) (model.Goal, error)
	EliminarMeta(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerEstadoMeta(context.Context, uuid.UUID, uuid.UUID) (EstadoMeta, error)
	ObtenerMeta(context.Context, uuid.UUID, uuid.UUID) (model.Goal, error)
	ObtenerMetas(context.Context, uuid.UUID) ([]model.Goal, error)
}

type HandlerMetas struct {
	service IServiceMetas
	errors  errores.RespuestaErrores
}

type EstadoMeta struct {
	MontoActual        int64   `json:"monto_actual"`
	PorcentajeActual   float64 `json:"porcentaje_actual"`
	MontoPendiente     int64   `json:"monto_pendiente"`
	PorcentajeFaltante float64 `json:"porcentaje_faltante"`
}

type PayloadCrearMeta struct {
	Nombre    string                 `json:"nombre" validate:"required,max=30"`
	Periodo   utilidades.PeriodoMeta `json:"periodo" validate:"required,oneof=dia semana mes"`
	MontoMeta int64                  `json:"monto_meta" validate:"required,gt=0"`
}

func NuevoHandlerMetas(s IServiceMetas, e errores.RespuestaErrores) *HandlerMetas {
	return &HandlerMetas{s, e}
}

func (h *HandlerMetas) Crear(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCrearMeta
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r)

	if empresaID == nil {
		h.errors.ErrorSolicitudInvalida(w, r, errors.New("no hay ID de empresa en la solicitud"))
		return
	}

	goal, err := h.service.CrearMeta(r.Context(), payload, empresaID.ID)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, goal); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerMetas) EliminarMeta(w http.ResponseWriter, r *http.Request) {
	metaID, err := uuid.Parse(chi.URLParam(r, "metaID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	if err := h.service.EliminarMeta(r.Context(), metaID, empresaID); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Meta eliminada"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerMetas) ObtenerEstadoMeta(w http.ResponseWriter, r *http.Request) {
	metaID, err := uuid.Parse(chi.URLParam(r, "metaID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	state, err := h.service.ObtenerEstadoMeta(r.Context(), metaID, empresaID)

	if err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, state); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerMetas) ObtenerMeta(w http.ResponseWriter, r *http.Request) {
	metaID, err := uuid.Parse(chi.URLParam(r, "metaID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	goal, err := h.service.ObtenerMeta(r.Context(), metaID, empresaID)

	if err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, goal); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerMetas) ObtenerMetas(w http.ResponseWriter, r *http.Request) {

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	goals, err := h.service.ObtenerMetas(r.Context(), empresaID)

	if err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, goals); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
