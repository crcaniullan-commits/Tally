package clavesacceso

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion, resend SendEmail) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenClavesAcceso(db)
	svc := NuevoServicioClavesAcceso(repo, resend)
	hdl := NuevoHandlerClavesAcceso(svc, err)

	r.Route("/claves", func(r chi.Router) {
		r.Post("/", middl.VerificarRol(utilidades.RolMunicipal, hdl.EmitirParaUsuario))
		r.Route("/IDAcceso", func(r chi.Router) {
			r.Post("/{claveID}", middl.VerificarRol(utilidades.RolMunicipal, hdl.Reenviar))
		})
		r.Route("/revocar", func(r chi.Router) {
			r.Post("/{codigo}", middl.VerificarRol(utilidades.RolMunicipal, hdl.Revocar))
		})
	})
}
