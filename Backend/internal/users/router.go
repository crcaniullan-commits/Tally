package users

import (
	"database/sql"

	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, tran *dbtx.Transactor, middl util.MiddlewareAuth) {
	err := errorhandler.NewErrorResponse(logger)
	redm := accesskeys.NewStorage(db)

	repo := NewStorage(db)
	svc := NewUserService(repo, redm, tran)
	hdl := NewUserHandler(svc, err)

	r.Route("/users", func(r chi.Router) {
		r.Patch("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Update))
		r.Delete("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Delete))
		r.Post("/exchange/{code}", middl.CheckOwnership(util.UserRoleUsuario, hdl.Exchange))
		r.Get("/municipal/{rut}", middl.CheckOwnership(util.UserRoleMunicipal, hdl.GetByRut))
	})
}
