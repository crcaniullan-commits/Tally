package categories

import (
	"context"
	"fmt"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceCategorias interface {
	Crear(context.Context, PayloadCategoria, *uuid.UUID) (model.Categorie, error)
	CambiarNombre(context.Context, PayloadCategoria, uuid.UUID, uuid.UUID) error
	BorrarCategoria(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerCategorias(context.Context, uuid.UUID) ([]model.Categorie, error)
}

type HandlerCategorias struct {
	service IServiceCategorias
	errors  errores.RespuestaErrores
}

func NuevoHandlerCategorias(s IServiceCategorias, e errores.RespuestaErrores) *HandlerCategorias {
	return &HandlerCategorias{s, e}
}

type PayloadCategoria struct {
	Nombre string `json:"nombre" validate:"required,max=30"`
}

func (h *HandlerCategorias) Crear(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCategoria
	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	categoria, err := h.service.Crear(r.Context(), payload, &empresa.ID)

	if err != nil {
		switch err {
		case ErrorCategoriaDuplicada:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, categoria); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerCategorias) ObtenerCategorias(w http.ResponseWriter, r *http.Request) {
	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	categorias, err := h.service.ObtenerCategorias(r.Context(), empresa.ID)
	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, categorias); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerCategorias) CambiarNombre(w http.ResponseWriter, r *http.Request) {
	categoriaID, err := uuid.Parse(chi.URLParam(r, "categoriaID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadCategoria
	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err := h.service.CambiarNombre(r.Context(), payload, categoriaID, empresa.ID); err != nil {
		switch err {
		case ErrorCategoriaDuplicada:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Nombre cambiado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}

func (h *HandlerCategorias) BorrarCategoria(w http.ResponseWriter, r *http.Request) {
	categoriaID, err := uuid.Parse(chi.URLParam(r, "categoria"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	empresa := utilidades.ObtenerEmpresaDelContexto(r)

	if empresa == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := h.service.BorrarCategoria(r.Context(), categoriaID, empresa.ID); err != nil {
		switch err {
		case ErrorCategoriaConMovimientos:
			h.errors.ErrorConflicto(w, r, err)
			return
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	mensaje := fmt.Sprintf("Categoría %v eliminada", categoriaID)

	if err := utilidades.RespuestaJSON(w, http.StatusOK, mensaje); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
