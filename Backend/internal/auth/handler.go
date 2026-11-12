package auth

import (
	"context"
	"net/http"

	clavesacceso "github.com/crcaniullan-commits/Tally/internal/clavesAcceso"
	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"golang.org/x/crypto/bcrypt"
)

type IServiceAuth interface {
	RegistrarUsuario(context.Context, *PayloadCrearUsuario) (string, error)
	Login(context.Context, *PayloadInicioSesion) (string, error)
	CrearMunicipal(context.Context, *PayloadCrearMunicipal) (string, error)
}

type HandlerAuth struct {
	service IServiceAuth
	errors  errores.RespuestaErrores
}

func NuevoHandlerAuth(s IServiceAuth, e errores.RespuestaErrores) *HandlerAuth {
	return &HandlerAuth{s, e}
}

type PayloadCrearUsuario struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Nombre   string `json:"nombre" validate:"required,max=100"`
	Rut      string `json:"rut" validate:"required,max=10"`
	Code     string `json:"code" validate:"required"`
}

// PayloadCrearMunicipal no incluye code: crear un municipal no implica canjear
// una llave de acceso.
type PayloadCrearMunicipal struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Nombre   string `json:"nombre" validate:"required,max=100"`
	Rut      string `json:"rut" validate:"required,min=9,max=12"`
}

func (h *HandlerAuth) Registrar(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCrearUsuario
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	data, err := h.service.RegistrarUsuario(r.Context(), &payload)

	if err != nil {
		switch err {
		case users.ErrDuplicateEmail:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		case users.ErrDuplicateRut:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		case clavesacceso.ErrorCodigoNoCanjeable:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, data); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}

type PayloadInicioSesion struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (h *HandlerAuth) Login(w http.ResponseWriter, r *http.Request) {
	var payload PayloadInicioSesion
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	data, err := h.service.Login(r.Context(), &payload)

	if err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		case bcrypt.ErrMismatchedHashAndPassword:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err = utilidades.RespuestaJSON(w, http.StatusAccepted, data); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}

func (h *HandlerAuth) CrearMunicipal(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCrearMunicipal
	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	data, err := h.service.CrearMunicipal(r.Context(), &payload)

	if err != nil {
		switch err {
		case users.ErrDuplicateEmail:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		case users.ErrDuplicateRut:
			h.errors.ErrorSolicitudInvalida(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, data); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

}
