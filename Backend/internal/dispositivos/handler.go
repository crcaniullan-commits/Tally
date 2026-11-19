package dispositivos

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

type IServiceDispositivos interface {
	Registrar(ctx context.Context, userID uuid.UUID, payload PayloadDispositivo) error
	Desvincular(ctx context.Context, userID uuid.UUID, token string) error
	Listar(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error)
}

type HandlerDispositivos struct {
	service IServiceDispositivos
	errors  errores.RespuestaErrores
}

func NuevoHandlerDispositivos(service IServiceDispositivos, errors errores.RespuestaErrores) *HandlerDispositivos {
	return &HandlerDispositivos{
		service: service,
		errors:  errors,
	}
}

type PayloadDispositivo struct {
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=android ios web"`
}

type PayloadEliminarDispositivo struct {
	Token string `json:"token" validate:"required"`
}

func (h *HandlerDispositivos) Registrar(w http.ResponseWriter, r *http.Request) {
	usuario := utilidades.ObtenerUsuarioDelContexto(r)
	if usuario == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	var payload PayloadDispositivo
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err := h.service.Registrar(r.Context(), usuario.ID, payload); err != nil {
		switch {
		case errors.Is(err, ErrorTokenRequerido), errors.Is(err, ErrorPlataformaInvalida):
			h.errors.ErrorSolicitudInvalida(w, r, err)
		default:
			h.errors.ErrorInterno(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HandlerDispositivos) Desvincular(w http.ResponseWriter, r *http.Request) {
	usuario := utilidades.ObtenerUsuarioDelContexto(r)
	if usuario == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" && r.Body != nil && r.ContentLength > 0 {
		var payload PayloadEliminarDispositivo
		if err := utilidades.LeerJSONYValidar(w, r, &payload); err == nil {
			token = payload.Token
		}
	}

	if token == "" {
		h.errors.ErrorSolicitudInvalida(w, r, ErrorTokenRequerido)
		return
	}

	if err := h.service.Desvincular(r.Context(), usuario.ID, token); err != nil {
		switch {
		case errors.Is(err, utilidades.ErrorNoEncontrado):
			h.errors.ErrorNoEncontrado(w, r, err)
		case errors.Is(err, ErrorTokenRequerido):
			h.errors.ErrorSolicitudInvalida(w, r, err)
		default:
			h.errors.ErrorInterno(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HandlerDispositivos) Listar(w http.ResponseWriter, r *http.Request) {
	usuario := utilidades.ObtenerUsuarioDelContexto(r)
	if usuario == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	dispositivos, err := h.service.Listar(r.Context(), usuario.ID)
	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, dispositivos); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
