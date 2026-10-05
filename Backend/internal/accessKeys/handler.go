package accesskeys

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

type ServiceAccessKey interface {
	IssueForUser(context.Context, uuid.UUID, string) (model.AccessKey, error)
	Resend(context.Context, uuid.UUID, string) error
	RevokePremature(context.Context, uuid.UUID, string) error
}

type AccessKeysHandler struct {
	service ServiceAccessKey
	errors  errorhandler.ErrorsResponse
}

func NewAccessKeysHandler(s ServiceAccessKey, e errorhandler.ErrorsResponse) *AccessKeysHandler {
	return &AccessKeysHandler{s, e}
}

type EmailPayload struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *AccessKeysHandler) IssueForUser(w http.ResponseWriter, r *http.Request) {
	var payload EmailPayload

	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	users := util.GetUserFromContext(r)

	accessKey, err := h.service.IssueForUser(r.Context(), users.ID, payload.Email)

	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.JsonResponse(w, http.StatusOK, accessKey); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *AccessKeysHandler) Resend(w http.ResponseWriter, r *http.Request) {
	accessKeyId, err := uuid.Parse(chi.URLParam(r, "keyID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	var payload EmailPayload

	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err := h.service.Resend(r.Context(), accessKeyId, payload.Email); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Correo reenviado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

func (h *AccessKeysHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if code == "" {
		h.errors.BadRequestResponse(w, r, errors.New("have to be a code"))
		return
	}

	users := util.GetUserFromContext(r)

	if err := h.service.RevokePremature(r.Context(), users.ID, code); err != nil {
		switch err {
		case util.ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusOK, "Codigo desabilitado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}
