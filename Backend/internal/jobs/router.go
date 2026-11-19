package jobs

import (
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ModuloJobs struct {
	Servicio     *ServicioJobs
	Planificador *PlanificadorJobs
}

func InitModule(
	r chi.Router,
	db *sql.DB,
	logger *zap.SugaredLogger,
	middl utilidades.MiddlewareAutenticacion,
	notifier Notificador,
	loc *time.Location,
	cronDiario, cronCobros string,
) (*ModuloJobs, error) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenJobs(db)
	svc := NuevoServicioJobs(repo, notifier, loc, logger)
	hdl := NuevoHandlerJobs(svc, err)

	planificador, cronErr := NuevoPlanificadorJobs(svc, loc, cronDiario, cronCobros, logger)
	if cronErr != nil {
		return nil, cronErr
	}

	r.Route("/jobs", func(r chi.Router) {
		r.Post("/recordatorio-diario", middl.VerificarRol(utilidades.RolUsuario, hdl.RecordatorioDiario))
		r.Post("/cobros-vencidos", middl.VerificarRol(utilidades.RolUsuario, hdl.CobrosVencidos))
		r.Post("/ejecutar-todos", middl.VerificarRol(utilidades.RolUsuario, hdl.EjecutarTodos))
	})

	return &ModuloJobs{
		Servicio:     svc,
		Planificador: planificador,
	}, nil
}
