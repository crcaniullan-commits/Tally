package empresas

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/middleware"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Routes struct {
	Collection func(r chi.Router)
	Item       func(r chi.Router)
}

func InitModule(db *sql.DB, logger *zap.SugaredLogger, middl util.MiddlewareAuth) (Routes, middleware.EmpresaFinder) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStorage(db)
	svc := NewEmpresaService(repo)
	hdl := NewEmpresasHandler(svc, err)

	return Routes{
		Collection: func(r chi.Router) {
			r.Post("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Create))
			r.Get("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.GetEmpresas))
		},
		Item: func(r chi.Router) {
			r.Patch("/{nombre}", middl.CheckOwnership(util.UserRoleUsuario, hdl.CambiarNombre))
			r.Delete("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.BorrarEmpresa))
		},
	}, repo
}
