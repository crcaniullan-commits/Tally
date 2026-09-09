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
	Create(context.Context, *CreateUserPaylaod) (*Users, error)
	Update(context.Context, uuid.UUID, *UpdateUserPayload) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, util.RUT) (*Users, error)
	GetByEmail(context.Context, string) (*Users, error)
}

type UsersHandler struct {
	service ServiceUsers
	errors  errorhandler.ErrorsResponse
}

type CreateUserPaylaod struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
	Rut      string `json:"rut" validate:"required,max=10"`
}

func NewUserHandler(s ServiceUsers, e errorhandler.ErrorsResponse) *UsersHandler {
	return &UsersHandler{s, e}
}

/*
*	Creación de usuario
* 	recibe un http.ResponseWritter y un pointer
+	a http.Request
*/
func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPaylaod
	if err := util.ReadJSON(w, r, &payload); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err := util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	user, err := h.service.Create(r.Context(), &payload)
	if err != nil {
		switch err {
		case ErrDuplicateEmail:
			h.errors.BadRequestResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err := util.JsonResponse(w, http.StatusCreated, user); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

type UpdateUserPayload struct {
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
}

/*
*	Actualiza un usuario
* 	recibe un http.ResponseWritter y un pointer
+	a http.Request
*/
func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	var payload UpdateUserPayload
	if err = util.ReadJSON(w, r, &payload); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}

	if err = util.Validate.Struct(payload); err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err = h.service.Update(r.Context(), userID, &payload); err != nil {
		switch err {
		case ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusOK, "Usuario actualizado con exito"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

/*
*	Elimina un usuario de la base de datos
* 	recibe un http.ResponseWritter y un pointer
+	a http.Request
*/
func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		h.errors.BadRequestResponse(w, r, err)
		return
	}

	if err = h.service.Delete(r.Context(), userID); err != nil {
		switch err {
		case ErrNotFound:
			h.errors.NotFoundResponse(w, r, err)
			return
		default:
			h.errors.InternalServerError(w, r, err)
			return
		}
	}

	if err = util.JsonResponse(w, http.StatusOK, "Usuario eliminado"); err != nil {
		h.errors.InternalServerError(w, r, err)
		return
	}
}

/*
*	Busca un usuario por email
* 	recibe un http.ResponseWritter y un pointer
+	a http.Request
*/
func (h *UsersHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	user, err := h.service.GetByEmail(r.Context(), email)

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

/*
*	Busca un usuario por su rut
* 	recibe un http.ResponseWritter y un pointer
+	a http.Request
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
