package users

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenUsuarios(db)
	svc := NuevoServicioUsuarios(repo)
	hdl := NuevoHandlerUsuarios(svc, err)

	r.Route("/usuarios", func(r chi.Router) {
		r.Patch("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Actualizar))
		r.Delete("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Eliminar))
		r.Get("/municipal/{rut}", middl.VerificarRol(utilidades.RolMunicipal, hdl.ObtenerPorRUT))
	})
}
