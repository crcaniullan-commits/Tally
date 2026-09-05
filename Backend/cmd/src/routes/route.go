package routes

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/cmd/src/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Routes struct {
	handler handler.Handler
}

func (router *Routes) Mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1", func(r chi.Router) {
		router.handler.Users.RegisterRoutes(r)
	})

	return r

}

func NewRoutes(h handler.Handler) *Routes {
	return &Routes{h}
}
