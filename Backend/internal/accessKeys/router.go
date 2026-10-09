package accesskeys

import (
	"database/sql"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl util.MiddlewareAuth, resend SendEmail) {
	err := errorhandler.NewErrorResponse(logger)

	repo := NewStorage(db)
	svc := NewAccessKeyService(repo, resend)
	hdl := NewAccessKeysHandler(svc, err)

	r.Route("/key", func(r chi.Router) {
		r.Post("/", middl.CheckOwnership(util.UserRoleMunicipal, hdl.IssueForUser))
		r.Route("/AccessID", func(r chi.Router) {
			r.Post("/{keyID}", middl.CheckOwnership(util.UserRoleMunicipal, hdl.Resend))
		})
		r.Route("/revoke", func(r chi.Router) {
			r.Post("/{code}", middl.CheckOwnership(util.UserRoleMunicipal, hdl.Revoke))
		})
	})
}
