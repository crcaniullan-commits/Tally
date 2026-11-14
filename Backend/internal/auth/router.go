package auth

import (
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	clavesacceso "github.com/crcaniullan-commits/Tally/internal/clavesAcceso"
	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	cSecret, cIss string
	cExp          time.Duration
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, tran *bdtx.Transactor,
	secret, iss string, exp time.Duration, email Remitente) {
	err := errores.NuevaRespuestaErrores(logger)
	rem := clavesacceso.NuevoAlmacenClavesAcceso(db)
	jwt := NewJWTAuthenticator(
		secret,
		iss,
		iss,
	)

	cSecret = secret
	cIss = iss
	cExp = exp

	repo := NuevoAlmacenAuth(db)
	svc := NuevoServicioAuth(repo, jwt, tran, rem, email)
	hdl := NuevoHandlerAuth(svc, err)

	r.Route("/autenticacion", func(r chi.Router) {
		r.Post("/", hdl.Registrar)
		r.Post("/inicio-sesion", hdl.Login)
		r.Post("/municipal", hdl.CrearMunicipal)
		r.Route("/recuperar-password", func(r chi.Router) {
			r.Post("/", hdl.SolicitarRecuperacion)
			r.Post("/verificar", hdl.VerificarCodigoRecuperacion)
			r.Post("/confirmar", hdl.ConfirmarRecuperacion)
		})
	})
}
