package users

import (
	"context"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceUsuarios interface {
	Actualizar(context.Context, *model.User, *PayloadActualizarUsuario) error
	Eliminar(context.Context, uuid.UUID) error
	ObtenerPorRUT(context.Context, utilidades.RUT) (*model.User, error)
}

type HandlerUsuarios struct {
	service IServiceUsuarios
	errors  errores.RespuestaErrores
}

func NuevoHandlerUsuarios(s IServiceUsuarios, e errores.RespuestaErrores) *HandlerUsuarios {
	return &HandlerUsuarios{s, e}
}

type PayloadActualizarUsuario struct {
	Password string `json:"password" validate:"omitempty,min=8,max=72"`
	Nombre   string `json:"nombre" validate:"omitempty,max=100"`
}

/*
*	handler para que el usuario actualize sus datos
 */
func (h *HandlerUsuarios) Actualizar(w http.ResponseWriter, r *http.Request) {
	var payload PayloadActualizarUsuario
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	user := utilidades.ObtenerUsuarioDelContexto(r)

	if err := h.service.Actualizar(r.Context(), user, &payload); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Usuario actualizado con éxito"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

/**
*	Handler para que el usuario elimine su cuenta
 */
func (h *HandlerUsuarios) Eliminar(w http.ResponseWriter, r *http.Request) {

	user := utilidades.ObtenerUsuarioDelContexto(r)

	if user == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	if err := h.service.Eliminar(r.Context(), user.ID); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Usuario eliminado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

/**
*	Handler para que los usuarios municipales
*	Buscen la cuenta de los usuarios emprendedores
 */
func (h *HandlerUsuarios) ObtenerPorRUT(w http.ResponseWriter, r *http.Request) {
	rut, err := utilidades.ParseRUT(chi.URLParam(r, "rut"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	user, err := h.service.ObtenerPorRUT(r.Context(), rut)

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

	if err = utilidades.RespuestaJSON(w, http.StatusOK, user); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}
