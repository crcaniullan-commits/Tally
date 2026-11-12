package intermediario

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func MiddlewaresGlobales() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.ClientIPFromHeader("X-Real-IP"),
		middleware.Logger,
		middleware.Recoverer,
	}
}
