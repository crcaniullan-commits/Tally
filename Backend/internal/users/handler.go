package users

import (
	"context"
	"errors"
	"net/http"

	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceUsers interface {
	Update(context.Context, *model.User, *UpdateUserPayload) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, util.RUT) (*model.User, error)
	ExchangeCode(context.Context, string, uuid.UUID) error
}

type UsersHandler struct {
	service ServiceUsers
	errors  errorhandler.ErrorsResponse
}

func NewUserHandler(s ServiceUsers, e errorhandler.ErrorsResponse) *UsersHandler {
	return &UsersHandler{s, e}
}

type UpdateUserPayload struct {
	Password string `json:"password" validate:"omitempty,min=8,max=72"`
	Name     string `json:"nombre" validate:"omitempty,max=100"`
}

/*
*	handler para que el usuario actualize sus datos
 */
func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	var payload UpdateUserPayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	user := util.GetUserFromContext(r)

	if err := h.service.Update(r.Context(), user, &payload); err != nil {
		switch err {
		case util.ErrNotFound:
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

	user := util.GetUserFromContext(r)

	if err := h.service.Delete(r.Context(), user.ID); err != nil {
		switch err {
		case util.ErrNotFound:
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
		case util.ErrNotFound:
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

func (h *UsersHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if code == "" {
		h.errors.BadRequestResponse(w, r, errors.New("Debe haber un codigo"))
		return
	}

	users := util.GetUserFromContext(r)

	if err := h.service.ExchangeCode(r.Context(), code, users.ID); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		case ErrCodeRedemed:
			h.errors.BadRequestResponse(w, r, err)
			return
		case accesskeys.ErrNotRedeemable:
			// El UPDATE de access_keys filtra las llaves canjeables (no canjeada,
			// sin vencer y no revocada): si no matchea ninguna fila, el código
			// está vencido, revocado o ya se usó.
			h.errors.BadRequestResponse(w, r, err)
			return
		case ErrCodeVencido, ErrCodeRevocado:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusAccepted, "Canjeado con exito, ¡Felicidades!"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
