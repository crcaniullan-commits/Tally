package clavesacceso

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

type IServiceClavesAcceso interface {
	EmitirParaUsuario(context.Context, uuid.UUID, string) (model.AccessKey, error)
	Reenviar(context.Context, uuid.UUID, string) error
	Revocar(context.Context, uuid.UUID, string) error
}

type HandlerClavesAcceso struct {
	service IServiceClavesAcceso
	errors  errores.RespuestaErrores
}

func NuevoHandlerClavesAcceso(s IServiceClavesAcceso, e errores.RespuestaErrores) *HandlerClavesAcceso {
	return &HandlerClavesAcceso{s, e}
}

type PayloadCorreo struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *HandlerClavesAcceso) EmitirParaUsuario(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCorreo

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	users := utilidades.ObtenerUsuarioDelContexto(r)

	accessKey, err := h.service.EmitirParaUsuario(r.Context(), users.ID, payload.Email)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, accessKey); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerClavesAcceso) Reenviar(w http.ResponseWriter, r *http.Request) {
	claveID, err := uuid.Parse(chi.URLParam(r, "claveID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadCorreo

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err := h.service.Reenviar(r.Context(), claveID, payload.Email); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Correo reenviado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerClavesAcceso) Revocar(w http.ResponseWriter, r *http.Request) {
	codigo := chi.URLParam(r, "codigo")

	if codigo == "" {
		h.errors.ErrorSolicitudInvalida(w, r, errors.New("debe ser un código"))
		return
	}

	users := utilidades.ObtenerUsuarioDelContexto(r)

	if err := h.service.Revocar(r.Context(), users.ID, codigo); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Código deshabilitado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
