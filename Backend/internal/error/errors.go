package errorhandler

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/cmd/src/httputil"
	"go.uber.org/zap"
)

type ErrorsResponse struct {
	logger *zap.SugaredLogger
}

func NewErrorResponse(l *zap.SugaredLogger) ErrorsResponse {
	return ErrorsResponse{l}
}

func (e *ErrorsResponse) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	httputil.WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (e *ErrorsResponse) ForbiddenResponse(w http.ResponseWriter, r *http.Request) {
	e.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	httputil.WriteJSONError(w, http.StatusForbidden, "forbidden")
}

func (e *ErrorsResponse) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	httputil.WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func (e *ErrorsResponse) ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	httputil.WriteJSONError(w, http.StatusConflict, err.Error())
}

func (e *ErrorsResponse) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	httputil.WriteJSONError(w, http.StatusNotFound, "not found")
}

func (e *ErrorsResponse) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (e *ErrorsResponse) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (e *ErrorsResponse) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	e.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	httputil.WriteJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
