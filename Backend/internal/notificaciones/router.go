package notificaciones

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(
	r chi.Router,
	db *sql.DB,
	logger *zap.SugaredLogger,
	middl utilidades.MiddlewareAutenticacion,
	dispositivos BuscadorDispositivos,
	transporte EnviadorPush,
) *ServicioNotificaciones {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenNotificaciones(db)
	svc := NuevoServicioNotificaciones(repo, transporte, dispositivos, logger)
	hdl := NuevoHandlerNotificaciones(svc, err)

	r.Route("/notificaciones", func(r chi.Router) {
		r.Get("/historial", middl.VerificarRol(utilidades.RolUsuario, hdl.ListarHistorial))
		r.Post("/probar", middl.VerificarRol(utilidades.RolUsuario, hdl.ProbarNotificacion))
	})

	return svc
}
