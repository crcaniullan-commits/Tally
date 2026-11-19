package notificaciones

import (
	"context"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type IServiceNotificaciones interface {
	Send(ctx context.Context, userID uuid.UUID, n model.Notification) error
	Listar(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error)
}

type HandlerNotificaciones struct {
	service IServiceNotificaciones
	errors  errores.RespuestaErrores
}

func NuevoHandlerNotificaciones(service IServiceNotificaciones, errors errores.RespuestaErrores) *HandlerNotificaciones {
	return &HandlerNotificaciones{
		service: service,
		errors:  errors,
	}
}

type PayloadPruebaNotificacion struct {
	Titulo string `json:"titulo" validate:"omitempty,max=100"`
	Cuerpo string `json:"cuerpo" validate:"omitempty,max=250"`
}

func (h *HandlerNotificaciones) ListarHistorial(w http.ResponseWriter, r *http.Request) {
	usuario := utilidades.ObtenerUsuarioDelContexto(r)
	if usuario == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	historial, err := h.service.Listar(r.Context(), usuario.ID)
	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, historial); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *HandlerNotificaciones) ProbarNotificacion(w http.ResponseWriter, r *http.Request) {
	usuario := utilidades.ObtenerUsuarioDelContexto(r)
	if usuario == nil {
		h.errors.ErrorNoAutorizado(w, r, utilidades.ErrorNoAutorizado)
		return
	}

	var payload PayloadPruebaNotificacion
	if r.Body != nil && r.ContentLength > 0 {
		_ = utilidades.LeerJSON(w, r, &payload)
	}

	titulo := payload.Titulo
	if titulo == "" {
		titulo = "Prueba de notificación"
	}
	cuerpo := payload.Cuerpo
	if cuerpo == "" {
		cuerpo = "¡Las notificaciones de Tally están funcionando correctamente!"
	}

	n := model.Notification{
		Title: titulo,
		Body:  cuerpo,
		Data: map[string]string{
			"tipo": "prueba",
		},
	}

	if err := h.service.Send(r.Context(), usuario.ID, n); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, "notificación de prueba enviada"); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}
