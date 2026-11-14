package errores

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"go.uber.org/zap"
)

type RespuestaErrores struct {
	logger *zap.SugaredLogger
}

func NuevaRespuestaErrores(l *zap.SugaredLogger) RespuestaErrores {
	return RespuestaErrores{l}
}

func (e *RespuestaErrores) ErrorInterno(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utilidades.EscribirErrorJSON(w, http.StatusInternalServerError, "el servidor encontró un problema")
}

func (e *RespuestaErrores) ErrorProhibido(w http.ResponseWriter, r *http.Request) {
	e.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	utilidades.EscribirErrorJSON(w, http.StatusForbidden, "prohibido")
}

func (e *RespuestaErrores) ErrorSolicitudInvalida(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utilidades.EscribirErrorJSON(w, http.StatusBadRequest, err.Error())
}

func (e *RespuestaErrores) ErrorConflicto(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utilidades.EscribirErrorJSON(w, http.StatusConflict, err.Error())
}

func (e *RespuestaErrores) ErrorNoEncontrado(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utilidades.EscribirErrorJSON(w, http.StatusNotFound, "no encontrado")
}

func (e *RespuestaErrores) ErrorNoAutorizado(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utilidades.EscribirErrorJSON(w, http.StatusUnauthorized, "no autorizado")
}

func (e *RespuestaErrores) ErrorNoAutorizadoBasico(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	utilidades.EscribirErrorJSON(w, http.StatusUnauthorized, "no autorizado")
}

func (e *RespuestaErrores) ErrorLimiteExcedido(w http.ResponseWriter, r *http.Request, retryAfter string) {
	e.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	utilidades.EscribirErrorJSON(w, http.StatusTooManyRequests, "límite de peticiones excedido, reintentar después de: "+retryAfter)
}
