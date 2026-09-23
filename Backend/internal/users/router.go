package users

import (
	"database/sql"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MiddlewareAuth interface {
	CheckOwnership(util.UserRole, http.HandlerFunc) http.HandlerFunc
}

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl MiddlewareAuth) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStorage(db)
	svc := NewUserService(repo)
	hdl := NewUserHandler(svc, err)

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Patch("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Update))
			r.Delete("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Delete))
		})
		r.Get("/municipal/{rut}", middl.CheckOwnership(util.UserRoleMunicipal, hdl.GetByRut))
	})
}
