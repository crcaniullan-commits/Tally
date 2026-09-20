package auth

import (
	"context"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
)

type ServiceAuth interface {
	RegisterUser(context.Context, *CreateUserPayload) (string, error)
	Login(context.Context, *LoginUserPayload) (string, error)
}

type AuthHandler struct {
	service ServiceAuth
	errors  errorhandler.ErrorsResponse
}

func NewAuthHandler(s ServiceAuth, e errorhandler.ErrorsResponse) *AuthHandler {
	return &AuthHandler{s, e}
}

type CreateUserPayload struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
	Rut      string `json:"rut" validate:"required,max=10"`
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	data, err := h.service.RegisterUser(r.Context(), &payload)

	if err != nil {
		switch err {
		case ErrDuplicateEmail:
			h.errors.BadRequestResponse(w, r, err)
			return
		case ErrDuplicateRut:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, data); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload LoginUserPayload
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	data, err := h.service.Login(r.Context(), &payload)

	if err != nil {
		switch err {
		case ErrNotFound:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusAccepted, data); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

}
