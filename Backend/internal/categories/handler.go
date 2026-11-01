package categories

import (
	"context"
	"fmt"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceCategories interface {
	Create(context.Context, CategoriaPayload, *uuid.UUID) (model.Categorie, error)
	CambiarNombre(context.Context, CategoriaPayload, uuid.UUID, uuid.UUID) error
	BorrarCategoria(context.Context, uuid.UUID, uuid.UUID) error
	GetCategorias(context.Context, uuid.UUID) ([]model.Categorie, error)
}

type CategoriesHandler struct {
	service ServiceCategories
	errors  errorhandler.ErrorsResponse
}

func NewCategoriesHandler(s ServiceCategories, e errorhandler.ErrorsResponse) *CategoriesHandler {
	return &CategoriesHandler{s, e}
}

type CategoriaPayload struct {
	Nombre string `json:"nombre" validate:"required,max=30"`
}

func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload CategoriaPayload
	empresa := util.GetEmpresaFromContext(r)

	if empresa == nil {
		h.errors.UnauthorizedErrorResponse(w, r, util.ErrUnauthorized)
		return
	}

	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	categoria, err := h.service.Create(r.Context(), payload, &empresa.ID)

	if err != nil {
		switch err {
		case AlreadyExistError:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusCreated, categoria); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *CategoriesHandler) GetCategorias(w http.ResponseWriter, r *http.Request) {
	empresa := util.GetEmpresaFromContext(r)

	if empresa == nil {
		h.errors.UnauthorizedErrorResponse(w, r, util.ErrUnauthorized)
		return
	}

	categorias, err := h.service.GetCategorias(r.Context(), empresa.ID)
	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, categorias); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *CategoriesHandler) CambiarNombre(w http.ResponseWriter, r *http.Request) {
	categoriaID, err := uuid.Parse(chi.URLParam(r, "categoriaID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	var payload CategoriaPayload
	empresa := util.GetEmpresaFromContext(r)

	if empresa == nil {
		h.errors.UnauthorizedErrorResponse(w, r, util.ErrUnauthorized)
		return
	}

	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := h.service.CambiarNombre(r.Context(), payload, categoriaID, empresa.ID); err != nil {
		switch err {
		case AlreadyExistError:
			h.errors.BadRequestResponse(w, r, err)
			return
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Nombre cambiado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}

func (h *CategoriesHandler) BorrarCategoria(w http.ResponseWriter, r *http.Request) {
	categoriaID, err := uuid.Parse(chi.URLParam(r, "categoria"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresa := util.GetEmpresaFromContext(r)

	if empresa == nil {
		h.errors.UnauthorizedErrorResponse(w, r, util.ErrUnauthorized)
		return
	}

	if err := h.service.BorrarCategoria(r.Context(), categoriaID, empresa.ID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	mensaje := fmt.Sprintf("Categoria %v eliminada", categoriaID)

	if err := util.JsonResponse(w, http.StatusOK, mensaje); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
