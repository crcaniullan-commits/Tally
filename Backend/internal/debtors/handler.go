package debtors

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceDeudores interface {
	Crear(context.Context, uuid.UUID, PayloadDeudor) (model.Debtor, error)
	Actualizar(context.Context, uuid.UUID, uuid.UUID, PayloadActualizacionDeudor) error
	MarcarDeudorPagado(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerDeudores(context.Context, uuid.UUID) ([]model.Debtor, error)
}

type HandlerDeudores struct {
	service IServiceDeudores
	error   errores.RespuestaErrores
}

func NuevoHandlerDeudores(s IServiceDeudores, e errores.RespuestaErrores) *HandlerDeudores {
	return &HandlerDeudores{s, e}
}

type PayloadDeudor struct {
	Nombre      string    `json:"nombre" validate:"required,max=30"`
	Monto       int64     `json:"monto" validate:"required,gt=0"`
	Descripcion string    `json:"descripcion" validate:"required,max=100"`
	FechaLimite time.Time `json:"fecha_limite" validate:"required"`
}

func (h *HandlerDeudores) Crear(w http.ResponseWriter, r *http.Request) {
	var payload PayloadDeudor

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.error.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	deb, err := h.service.Crear(r.Context(), empresaID, payload)

	if err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, deb); err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}
}

type PayloadActualizacionDeudor struct {
	Nombre      string    `json:"nombre" validate:"required,max=30"`
	Monto       int64     `json:"monto" validate:"required,gt=0"`
	FechaLimite time.Time `json:"fecha_limite" validate:"required"`
}

func (h *HandlerDeudores) Actualizar(w http.ResponseWriter, r *http.Request) {
	deudorID, err := uuid.Parse(chi.URLParam(r, "deudorID"))

	if err != nil {
		h.error.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadActualizacionDeudor

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.error.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	if err := h.service.Actualizar(r.Context(), deudorID, empresaID, payload); err != nil {
		switch {
		case errors.Is(err, utilidades.ErrorNoEncontrado):
			h.error.ErrorNoEncontrado(w, r, err)
		default:
			h.error.ErrorInterno(w, r, err)
		}
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Deuda actualizada"); err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerDeudores) MarcarDeudorPagado(w http.ResponseWriter, r *http.Request) {
	deudorID, err := uuid.Parse(chi.URLParam(r, "deudorID"))

	if err != nil {
		h.error.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	if err := h.service.MarcarDeudorPagado(r.Context(), deudorID, empresaID); err != nil {
		switch {
		case errors.Is(err, utilidades.ErrorNoEncontrado):
			h.error.ErrorNoEncontrado(w, r, err)
		default:
			h.error.ErrorInterno(w, r, err)
		}
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Deuda pagada"); err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerDeudores) ObtenerDeudoresDeEmpresa(w http.ResponseWriter, r *http.Request) {
	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	debtors, err := h.service.ObtenerDeudores(r.Context(), empresaID)
	if err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, debtors); err != nil {
		h.error.ErrorInterno(w, r, err)
		return
	}
}
