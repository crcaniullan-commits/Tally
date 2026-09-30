package incomes

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
	svc := NewIncomeService(repo)
	hdl := NewIncomesHandler(svc, err)

	r.Route("/incomes", func(r chi.Router) {
		r.Post("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.AddIncome))
		r.Delete("/{incomeID}", middl.CheckOwnership(util.UserRoleUsuario, hdl.Delete))
		r.Get("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.GetIncomesOfUser))
	})
}
