package incomes

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

type IServiceIngresos interface {
	AgregarIngreso(context.Context, PayloadIngreso, uuid.UUID) (model.Income, error)
	ActualizarIngreso(context.Context, PayloadActualizarIngreso, uuid.UUID, uuid.UUID) (model.Income, error)
	EliminarIngreso(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerIngresosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Income, error)
}

type HandlerIngresos struct {
	service IServiceIngresos
	errors  errores.RespuestaErrores
}

func NuevoHandlerIngresos(s IServiceIngresos, e errores.RespuestaErrores) *HandlerIngresos {
	return &HandlerIngresos{s, e}
}

func empresaID(r *http.Request) uuid.UUID {
	return utilidades.ObtenerEmpresaDelContexto(r).ID
}

type PayloadIngreso struct {
	Monto         int64               `json:"monto" validate:"required,gt=0"`
	PaymentMethod model.PaymentMethod `json:"payment_method" validate:"required,oneof=debito credito transferencia efectivo"`
	Descripcion   *string             `json:"descripcion" validate:"omitempty,max=100"`
	CategoryID    *uuid.UUID          `json:"category_id" validate:"omitempty"`
}

// PayloadActualizarIngreso son los campos editables de un ingreso. Lo que no
// viene en el payload se queda como está; "fecha" (AAAA-MM-DD) permite
// corregir la fecha del movimiento.
type PayloadActualizarIngreso struct {
	Monto         *int64               `json:"monto" validate:"omitempty,gt=0"`
	PaymentMethod *model.PaymentMethod `json:"payment_method" validate:"omitempty,oneof=debito credito transferencia efectivo"`
	Descripcion   *string              `json:"descripcion" validate:"omitempty,max=100"`
	CategoryID    *uuid.UUID           `json:"category_id"`
	Fecha         *string              `json:"fecha"`
}

/*
*	Añade un nuevo ingreso
 */
func (h *HandlerIngresos) AgregarIngreso(w http.ResponseWriter, r *http.Request) {
	var payload PayloadIngreso
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	income, err := h.service.AgregarIngreso(r.Context(), payload, empresaID(r))

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

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, income); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}

/*
*	Actualiza un ingreso: monto, método de pago, descripción, categoría y fecha
 */
func (h *HandlerIngresos) Actualizar(w http.ResponseWriter, r *http.Request) {
	ingresoID, err := uuid.Parse(chi.URLParam(r, "ingresoID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadActualizarIngreso

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	income, err := h.service.ActualizarIngreso(r.Context(), payload, ingresoID, empresaID(r))

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

	if err := utilidades.RespuestaJSON(w, http.StatusOK, income); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

/*
*	Elimina un ingreso
 */
func (h *HandlerIngresos) Eliminar(w http.ResponseWriter, r *http.Request) {
	ingresoID, err := uuid.Parse(chi.URLParam(r, "ingresoID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err = h.service.EliminarIngreso(r.Context(), ingresoID, empresaID(r)); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err = utilidades.RespuestaJSON(w, http.StatusOK, "ingreso eliminado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

/*
*	Obtiene todos los ingresos de la empresa
*
*	Acepta los query params de paginación: limite (1..30, default 30), desplazamiento y
*	el rango de fechas desde/hasta (AAAA-MM-DD). Un parámetro mal formado es un
*	400, no se cae al default en silencio.
 */
func (h *HandlerIngresos) ObtenerIngresosDeEmpresa(w http.ResponseWriter, r *http.Request) {
	fq, err := paginacion.NuevaConsultaPaginacionIngresos().Parse(r)

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	incomes, err := h.service.ObtenerIngresosDeEmpresa(r.Context(), empresaID(r), fq)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, incomes); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}
