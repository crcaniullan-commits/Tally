package goals

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenMetas(db)
	svc := NuevoServicioMetas(repo)
	hdl := NuevoHandlerMetas(svc, err)

	r.Route("/metas", func(r chi.Router) {
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Crear))
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerMetas))
		r.Route("/{metaID}", func(r chi.Router) {
			r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerMeta))
			r.Delete("/", middl.VerificarRol(utilidades.RolUsuario, hdl.EliminarMeta))
			r.Get("/estado", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerEstadoMeta))
		})
	})
}
