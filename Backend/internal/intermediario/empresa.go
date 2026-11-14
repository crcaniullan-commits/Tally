package intermediario

import (
	"context"
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BuscadorEmpresa interface {
	ObtenerPorID(context.Context, uuid.UUID, uuid.UUID) (*model.Empresa, error)
}
type MiddlewareEmpresa struct {
	error errores.RespuestaErrores
	find  BuscadorEmpresa
}

func NuevoMiddlewareEmpresa(logger *zap.SugaredLogger, f BuscadorEmpresa) *MiddlewareEmpresa {
	err := errores.NuevaRespuestaErrores(logger)
	return &MiddlewareEmpresa{err, f}
}

func (m *MiddlewareEmpresa) EmpresaDelUsuario(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := uuid.Parse(chi.URLParam(r, "empresaID"))

		if err != nil {
			m.error.ErrorSolicitudInvalida(w, r, err)
			return
		}

		userID := utilidades.ObtenerUsuarioDelContexto(r).ID

		empresa, err := m.find.ObtenerPorID(r.Context(), empresaID, userID)

		if err != nil {
			switch err {
			case utilidades.ErrorNoEncontrado:
				m.error.ErrorNoEncontrado(w, r, err)
				return
			default:
				m.error.ErrorNoAutorizado(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), utilidades.ContextoEmpresa, empresa)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
