package users

import (
	"context"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceUsers interface {
	Update(context.Context, *Users, *UpdateUserPayload) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, util.RUT) (*Users, error)
}

type UsersHandler struct {
	service ServiceUsers
	errors  errorhandler.ErrorsResponse
}

func NewUserHandler(s ServiceUsers, e errorhandler.ErrorsResponse) *UsersHandler {
	return &UsersHandler{s, e}
}

type UpdateUserPayload struct {
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
}

/*
*	handler para que el usuario actualize sus datos
 */
func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	var payload UpdateUserPayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	user := GetUserFromContext(r)

	if err := h.service.Update(r.Context(), user, &payload); err != nil {
		switch err {
		case ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Usuario actualizado con exito"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

/**
*	Handler para que el usuario elimine su cuenta
 */
func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {

	user := GetUserFromContext(r)

	if err := h.service.Delete(r.Context(), user.ID); err != nil {
		switch err {
		case ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Usuario eliminado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

/**
*	Handler para que los usuarios municipales
*	Buscen la cuenta de los usuarios emprendedores
 */
func (h *UsersHandler) GetByRut(w http.ResponseWriter, r *http.Request) {
	rut, err := util.ParseRUT(chi.URLParam(r, "rut"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
	}

	user, err := h.service.GetByRut(r.Context(), rut)

	if err != nil {
		switch err {
		case ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusOK, user); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}
