package dispositivos

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) *ServicioDispositivos {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenDispositivos(db)
	svc := NuevoServicioDispositivos(repo)
	hdl := NuevoHandlerDispositivos(svc, err)

	r.Route("/dispositivos", func(r chi.Router) {
		r.Put("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Registrar))
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Registrar))
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Listar))
		r.Delete("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Desvincular))
		r.Delete("/{token}", middl.VerificarRol(utilidades.RolUsuario, hdl.Desvincular))
	})

	return svc
}
