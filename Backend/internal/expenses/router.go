package expenses

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenGastos(db)
	svc := NuevoServicioGastos(repo)
	hdl := NuevoHandlerGastos(svc, err)

	r.Route("/gastos", func(r chi.Router) {
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Crear))
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerGastosDeEmpresa))
		r.Delete("/{gastoID}", middl.VerificarRol(utilidades.RolUsuario, hdl.Eliminar))
	})
}
