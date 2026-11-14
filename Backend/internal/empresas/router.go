package empresas

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/intermediario"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Routes struct {
	Collection func(r chi.Router)
	Item       func(r chi.Router)
}

func InitModule(db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) (Routes, intermediario.BuscadorEmpresa) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenEmpresas(db)
	svc := NuevoServicioEmpresas(repo)
	hdl := NuevoHandlerEmpresas(svc, err)

	return Routes{
		Collection: func(r chi.Router) {
			r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Crear))
			r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerEmpresas))
		},
		Item: func(r chi.Router) {
			r.Patch("/{nombre}", middl.VerificarRol(utilidades.RolUsuario, hdl.CambiarNombre))
			r.Delete("/", middl.VerificarRol(utilidades.RolUsuario, hdl.BorrarEmpresa))
		},
	}, repo
}
