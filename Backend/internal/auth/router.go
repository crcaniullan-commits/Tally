package auth

import (
	"database/sql"
	"time"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	cSecret, cIss string
	cExp          time.Duration
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger,
	secret, iss string, exp time.Duration) {
	err := errorhandler.NewErrorResponse(logger)
	jwt := NewJWTAuthenticator(
		secret,
		iss,
		iss,
	)

	cSecret = secret
	cIss = iss
	cExp = exp

	repo := NewStorage(db)
	svc := NewAuthService(repo, jwt)
	hdl := NewAuthHandler(svc, err)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/", hdl.RegisterHandler)
		r.Post("/login", hdl.Login)
	})
}
