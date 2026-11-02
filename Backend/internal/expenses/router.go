package expenses

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl util.MiddlewareAuth) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStorage(db)
	svc := NewExpenseService(repo)
	hdl := NewExpensesHandler(svc, err)

	r.Route("/expenses", func(r chi.Router) {
		r.Post("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Create))
		r.Get("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.GetExpensesOfEmpresa))
		r.Delete("/{expenseID}", middl.CheckOwnership(util.UserRoleUsuario, hdl.Delete))
	})
}
