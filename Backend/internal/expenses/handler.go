package expenses

import (
	"context"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceGastos interface {
	AgregarGasto(context.Context, PayloadGasto, uuid.UUID) (model.Expense, error)
	ActualizarGasto(context.Context, PayloadActualizarGasto, uuid.UUID, uuid.UUID) (model.Expense, error)
	EliminarGasto(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerGastosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error)
}

type HandlerGastos struct {
	service IServiceGastos
	errors  errores.RespuestaErrores
}

func NuevoHandlerGastos(s IServiceGastos, e errores.RespuestaErrores) *HandlerGastos {
	return &HandlerGastos{s, e}
}

func empresaID(r *http.Request) uuid.UUID {
	return utilidades.ObtenerEmpresaDelContexto(r).ID
}

type PayloadGasto struct {
	Monto       int64      `json:"monto" validate:"required,gt=0"`
	CategoryID  *uuid.UUID `json:"category_id" validate:"required"`
	Descripcion *string    `json:"descripcion" validate:"omitempty,max=100"`
}

// PayloadActualizarGasto son los campos editables de un gasto. Lo que no viene
// en el payload se queda como está; "fecha" (AAAA-MM-DD) permite corregir la
// fecha del movimiento.
type PayloadActualizarGasto struct {
	Monto       *int64     `json:"monto" validate:"omitempty,gt=0"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Descripcion *string    `json:"descripcion" validate:"omitempty,max=100"`
	Fecha       *string    `json:"fecha"`
}

// Actualizar edita un gasto: monto, categoría, descripción y fecha.
func (h *HandlerGastos) Actualizar(w http.ResponseWriter, r *http.Request) {
	gastoID, err := uuid.Parse(chi.URLParam(r, "gastoID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadActualizarGasto

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	expense, err := h.service.ActualizarGasto(r.Context(), payload, gastoID, empresaID(r))

	if err != nil {
		switch err {
		case ErrorMovimientoNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		case ErrorFechaInvalida, ErrorSinCambios, utilidades.ErrorNoEncontrado:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, expense); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerGastos) Crear(w http.ResponseWriter, r *http.Request) {
	var payload PayloadGasto
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	expense, err := h.service.AgregarGasto(r.Context(), payload, empresaID(r))

	if err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, expense); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerGastos) Eliminar(w http.ResponseWriter, r *http.Request) {
	gastoID, err := uuid.Parse(chi.URLParam(r, "gastoID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err = h.service.EliminarGasto(r.Context(), gastoID, empresaID(r)); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err = utilidades.RespuestaJSON(w, http.StatusOK, "gasto eliminado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerGastos) ObtenerGastosDeEmpresa(w http.ResponseWriter, r *http.Request) {
	fq, err := paginacion.NuevaConsultaPaginacionIngresos().Parse(r)

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	expenses, err := h.service.ObtenerGastosDeEmpresa(r.Context(), empresaID(r), fq)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, expenses); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
