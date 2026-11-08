package goals

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl util.MiddlewareAuth) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStoreGoal(db)
	svc := NewGoalService(repo)
	hdl := NewGoalsHandler(svc, err)

	r.Route("/goal", func(r chi.Router) {
		r.Post("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.Create))
		r.Get("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.getGoals))
		r.Route("/{goalID}", func(r chi.Router) {
			r.Get("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.getGoal))
			r.Delete("/", middl.CheckOwnership(util.UserRoleUsuario, hdl.DeleteGoal))
			r.Get("/state", middl.CheckOwnership(util.UserRoleUsuario, hdl.GetGoalState))
		})
	})
}
