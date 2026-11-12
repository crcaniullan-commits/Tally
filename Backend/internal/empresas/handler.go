package empresas

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

type IServiceEmpresas interface {
	Crear(context.Context, PayloadEmpresa) (model.Empresa, error)
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarEmpresa(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerEmpresas(context.Context, uuid.UUID) ([]model.Empresa, error)
}

type HandlerEmpresas struct {
	service IServiceEmpresas
	errors  errores.RespuestaErrores
}

func NuevoHandlerEmpresas(s IServiceEmpresas, e errores.RespuestaErrores) *HandlerEmpresas {
	return &HandlerEmpresas{s, e}
}

type PayloadEmpresa struct {
	NombreEmpresa string `json:"nombre_empresa" validate:"required,max=30"`
	userID        uuid.UUID
}

func (h *HandlerEmpresas) Crear(w http.ResponseWriter, r *http.Request) {
	var payload PayloadEmpresa
	user := utilidades.ObtenerUsuarioDelContexto(r)

	if user == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	payload.userID = user.ID

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresa, err := h.service.Crear(r.Context(), payload)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, empresa); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerEmpresas) ObtenerEmpresas(w http.ResponseWriter, r *http.Request) {
	user := utilidades.ObtenerUsuarioDelContexto(r)

	if user == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	empresas, err := h.service.ObtenerEmpresas(r.Context(), user.ID)

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

	if err := utilidades.RespuestaJSON(w, http.StatusOK, empresas); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerEmpresas) CambiarNombre(w http.ResponseWriter, r *http.Request) {
	nombre := chi.URLParam(r, "nombre")

	if nombre == "" {
		h.errors.ErrorSolicitudInvalida(w, r, errors.New("la ruta debe incluir el nombre de la empresa"))
		return
	}

	user := utilidades.ObtenerUsuarioDelContexto(r)
	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if user == nil || empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := h.service.CambiarNombre(r.Context(), nombre, empresa.ID, user.ID); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Nombre actualizado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerEmpresas) BorrarEmpresa(w http.ResponseWriter, r *http.Request) {
	user := utilidades.ObtenerUsuarioDelContexto(r)
	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if user == nil || empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := h.service.BorrarEmpresa(r.Context(), empresa.ID, user.ID); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "empresa borrada"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
