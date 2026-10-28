package empresas

import (
	"context"
	"errors"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceEmpresas interface {
	Create(context.Context, empresaPayload) (model.Empresa, error)
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarEmpresa(context.Context, uuid.UUID, uuid.UUID) error
	GetEmpresas(context.Context, uuid.UUID) ([]model.Empresa, error)
}

type EmpresasHandler struct {
	service ServiceEmpresas
	errors  errorhandler.ErrorsResponse
}

func NewEmpresasHandler(s ServiceEmpresas, e errorhandler.ErrorsResponse) *EmpresasHandler {
	return &EmpresasHandler{s, e}
}

type empresaPayload struct {
	Nombre_empresa string `json:"nombre_empresa" validate:"required,max=30"`
	user_id        uuid.UUID
}

func (h *EmpresasHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload empresaPayload
	userID := util.GetUserFromContext(r)

	payload.user_id = userID.ID

	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	empresa, err := h.service.Create(r.Context(), payload)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusCreated, empresa); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *EmpresasHandler) GetEmpresas(w http.ResponseWriter, r *http.Request) {
	userID := util.GetUserFromContext(r).ID

	empresas, err := h.service.GetEmpresas(r.Context(), userID)

	if err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, empresas); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *EmpresasHandler) CambiarNombre(w http.ResponseWriter, r *http.Request) {
	nombre := chi.URLParam(r, "nombre")

	if nombre == "" {
		h.errors.BadRequestResponse(w, r, errors.New("la ruta debe incluir el nombre de la empresa"))
		return
	}

	userID := util.GetUserFromContext(r).ID
	empresaID := util.GetEmpresaFromContext(r).ID

	if err := h.service.CambiarNombre(r.Context(), nombre, empresaID, userID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Nombre acualizado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *EmpresasHandler) BorrarEmpresa(w http.ResponseWriter, r *http.Request) {
	userID := util.GetUserFromContext(r).ID
	empresaID := util.GetEmpresaFromContext(r).ID

	if err := h.service.BorrarEmpresa(r.Context(), empresaID, userID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "empresa borrada"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
