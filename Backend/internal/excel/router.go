package excel

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// InitModule monta las rutas de excel. Se llama dentro del grupo de la
// empresa (/empresas/{empresaID}), igual que ingresos, gastos, categorias,
// metas y deudas: la empresa de la que salen los archivos es la del contexto,
// no la del path.
func InitModule(r chi.Router, db *sql.DB, fs FileStorage, logger *zap.SugaredLogger, mddl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)
	svc := &Service{store: &Store{db: db}, storage: fs}
	h := NuevoHandlerExcel(svc, err)

	r.Route("/excel", func(r chi.Router) {
		r.Post("/", mddl.VerificarRol(utilidades.RolUsuario, h.Upload))
		r.Get("/", mddl.VerificarRol(utilidades.RolUsuario, h.List))
		r.Get("/{id}", mddl.VerificarRol(utilidades.RolUsuario, h.Download))
		r.Delete("/{id}", mddl.VerificarRol(utilidades.RolUsuario, h.Delete))
	})

}
