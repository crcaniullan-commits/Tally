package application

import (
	"database/sql"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/categories"
	clavesacceso "github.com/crcaniullan-commits/Tally/internal/clavesAcceso"
	"github.com/crcaniullan-commits/Tally/internal/debtors"
	"github.com/crcaniullan-commits/Tally/internal/empresas"
	"github.com/crcaniullan-commits/Tally/internal/excel"
	"github.com/crcaniullan-commits/Tally/internal/expenses"
	"github.com/crcaniullan-commits/Tally/internal/goals"
	"github.com/crcaniullan-commits/Tally/internal/incomes"
	"github.com/crcaniullan-commits/Tally/internal/intermediario"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

func AsignarRutas(r chi.Router, smtp remitidorCorreo, transactor *bdtx.Transactor, empresaRoute empresas.Routes, emMiddl intermediario.MiddlewareEmpresa,
	fileStore excel.FileStorage, db *sql.DB, logger *zap.SugaredLogger, Addr string, jwt ConfigJWT, middl *intermediario.MiddlewareAutenticacion,
	buscadorUsuario users.BuscadorUsuarioConCache) func(r chi.Router) {
	return func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/swagger/doc.json", Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		auth.InitModule(r, db, logger, transactor,
			jwt.Secret, jwt.Iss, jwt.Exp, smtp)
		r.Route("/aplicacion", func(r chi.Router) {
			r.Use(middl.MiddlewareTokenAutenticacion)
			users.InitModule(r, db, logger, middl, &buscadorUsuario)
			clavesacceso.InitModule(r, db, logger, middl, smtp)
			r.Route("/empresas", func(r chi.Router) {
				empresaRoute.Collection(r)

				r.Route("/{empresaID}", func(r chi.Router) {
					r.Use(emMiddl.EmpresaDelUsuario)

					empresaRoute.Item(r)
					incomes.InitModule(r, db, logger, middl)
					expenses.InitModule(r, db, logger, middl)
					categories.InitModule(r, db, logger, middl)
					goals.InitModule(r, db, logger, middl)
					debtors.InitModule(r, db, logger, middl)
					excel.InitModule(r, db, fileStore, logger, middl)
				})
			})
		})
	}
}
