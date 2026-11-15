package clavesacceso

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IServiceClavesAcceso interface {
	EmitirParaUsuario(context.Context, uuid.UUID, string) (model.AccessKey, error)
	Reenviar(context.Context, uuid.UUID, string) error
	Revocar(context.Context, uuid.UUID, string) error
	ListarPorEmisor(context.Context, uuid.UUID, int, int) ([]model.AccessKey, error)
}

type HandlerClavesAcceso struct {
	service IServiceClavesAcceso
	errors  errores.RespuestaErrores
}

func NuevoHandlerClavesAcceso(s IServiceClavesAcceso, e errores.RespuestaErrores) *HandlerClavesAcceso {
	return &HandlerClavesAcceso{s, e}
}

type PayloadCorreo struct {
	Email string `json:"email" validate:"required,email"`
}

const (
	// limitePorDefecto es el tamaño de página del listado de claves cuando el
	// cliente no manda ?limite.
	limitePorDefecto = 30
	// limiteMaximo es el tope duro de ?limite. Un limite más grande se
	// responde con 400 en vez de recortar en silencio.
	limiteMaximo = 100
)

// ConsultaPaginacionClaves son los query params de GET /aplicacion/claves:
// ?limite y ?desplazamiento.
type ConsultaPaginacionClaves struct {
	Limite         int
	Desplazamiento int
}

// Parse arma la paginación desde la query string. Un parámetro inválido
// devuelve error en vez de ignorarlo en silencio: si ?limite=abc se comiera,
// el cliente creería que pidió una página y recibiría otra.
func (cq ConsultaPaginacionClaves) Parse(r *http.Request) (ConsultaPaginacionClaves, error) {
	qs := r.URL.Query()

	if v := qs.Get("limite"); v != "" {
		limite, err := strconv.Atoi(v)
		if err != nil {
			return cq, fmt.Errorf("limite inválido: %q no es un entero", v)
		}
		cq.Limite = limite
	}

	if v := qs.Get("desplazamiento"); v != "" {
		desplazamiento, err := strconv.Atoi(v)
		if err != nil {
			return cq, fmt.Errorf("desplazamiento inválido: %q no es un entero", v)
		}
		cq.Desplazamiento = desplazamiento
	}

	if cq.Limite < 1 || cq.Limite > limiteMaximo {
		return cq, fmt.Errorf("limite inválido: debe estar entre 1 y %d", limiteMaximo)
	}

	if cq.Desplazamiento < 0 {
		return cq, fmt.Errorf("desplazamiento inválido: no puede ser negativo")
	}

	return cq, nil
}

// Listar devuelve las claves de acceso emitidas por el municipal autenticado.
//
// Solo se listan las del municipal del contexto: la ruta es de uso interno del
// funcionario y no expone las claves emitidas por otros.
func (h *HandlerClavesAcceso) Listar(w http.ResponseWriter, r *http.Request) {
	consulta, err := ConsultaPaginacionClaves{Limite: limitePorDefecto}.Parse(r)
	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	users := utilidades.ObtenerUsuarioDelContexto(r)

	claves, err := h.service.ListarPorEmisor(r.Context(), users.ID, consulta.Limite, consulta.Desplazamiento)
	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, claves); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerClavesAcceso) EmitirParaUsuario(w http.ResponseWriter, r *http.Request) {
	var payload PayloadCorreo

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	users := utilidades.ObtenerUsuarioDelContexto(r)

	accessKey, err := h.service.EmitirParaUsuario(r.Context(), users.ID, payload.Email)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusCreated, accessKey); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerClavesAcceso) Reenviar(w http.ResponseWriter, r *http.Request) {
	claveID, err := uuid.Parse(chi.URLParam(r, "claveID"))

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	var payload PayloadCorreo

	if err := utilidades.LeerJSONYValidar(w, r, &payload); err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	if err := h.service.Reenviar(r.Context(), claveID, payload.Email); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Correo reenviado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerClavesAcceso) Revocar(w http.ResponseWriter, r *http.Request) {
	codigo := chi.URLParam(r, "codigo")

	if codigo == "" {
		h.errors.ErrorSolicitudInvalida(w, r, errors.New("debe ser un código"))
		return
	}

	users := utilidades.ObtenerUsuarioDelContexto(r)

	if err := h.service.Revocar(r.Context(), users.ID, codigo); err != nil {
		switch err {
		case utilidades.ErrorNoEncontrado:
			h.errors.ErrorNoEncontrado(w, r, err)
			return
		default:
			h.errors.ErrorInterno(w, r, err)
			return
		}
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "Código deshabilitado"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
