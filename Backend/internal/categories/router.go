package categories

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl util.MiddlewareAuth) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStoreCategorie(db)
	svc := NewCategorieService(repo)
	hdl := NewCategoriesHandler(svc, err)

	r.Route("/categories", func(r chi.Router) {
		r.Post("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Create))
		r.Patch("/{categoriaID}", middl.CheckOwnership(util.UserRoleUsuario, hdl.CambiarNombre))
		r.Delete("/{categoria}", middl.CheckOwnership(util.UserRoleUsuario, hdl.BorrarCategoria))
	})
}
