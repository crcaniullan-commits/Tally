package categories

import (
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func InitModule(r chi.Router, db *sql.DB, logger *zap.SugaredLogger, middl utilidades.MiddlewareAutenticacion) {
	err := errores.NuevaRespuestaErrores(logger)

	repo := NuevoAlmacenCategorias(db)
	svc := NuevoServicioCategorias(repo)
	hdl := NuevoHandlerCategorias(svc, err)

	r.Route("/categorias", func(r chi.Router) {
		r.Get("/", middl.VerificarRol(utilidades.RolUsuario, hdl.ObtenerCategorias))
		r.Post("/", middl.VerificarRol(utilidades.RolUsuario, hdl.Crear))
		r.Patch("/{categoriaID}", middl.VerificarRol(utilidades.RolUsuario, hdl.CambiarNombre))
		r.Delete("/{categoria}", middl.VerificarRol(utilidades.RolUsuario, hdl.BorrarCategoria))
	})
}
