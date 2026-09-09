package users

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStorage(db)
	svc := NewUserService(repo)
	hdl := NewUserHandler(svc, err)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", hdl.Create)
		r.Route("/{userID}", func(r chi.Router) {
			r.Patch("/", hdl.Update)
			r.Delete("/", hdl.Delete)
		})
		r.Get("/{email}", hdl.GetByEmail)
		r.Get("/{rut}", hdl.GetByRut)
	})
}
