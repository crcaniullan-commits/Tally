package debtors

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenDeudores(db)
	svc := NuevoServicioDeudores(repo)
	hdl := NuevoHandlerDeudores(svc, err)

	r.Route("/deudas", func(r chi.Router) {
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Crear))
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerDeudoresDeEmpresa))
		r.Patch("/{deudorID}", middl.VerificarRol(utilidades.RolUsuario, hdl.Actualizar))
		r.Patch("/{deudorID}/pagada", middl.VerificarRol(utilidades.RolUsuario, hdl.MarcarDeudorPagado))
	})
}
