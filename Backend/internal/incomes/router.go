package incomes

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenIngresos(db)
	svc := NuevoServicioIngresos(repo)
	hdl := NuevoHandlerIngresos(svc, err)

	r.Route("/ingresos", func(r chi.Router) {
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.AgregarIngreso))
		r.Patch("/{ingresoID}", middl.VerificarRol(utilidades.RolUsuario, hdl.Actualizar))
		r.Delete("/{ingresoID}", middl.VerificarRol(utilidades.RolUsuario, hdl.Eliminar))
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerIngresosDeEmpresa))
	})
}
