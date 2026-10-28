package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmpresaFinder interface {
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*model.Empresa, error)
}
type EmpresaMiddleware struct {
	error errorhandler.ErrorsResponse
	find  EmpresaFinder
}

func NewEmpresaMiddleware(logger *zap.SugaredLogger, f EmpresaFinder) *EmpresaMiddleware {
	err := errorhandler.NewErrorResponse(logger)
	return &EmpresaMiddleware{err, f}
}

func (m *EmpresaMiddleware) EmpresaOfUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := uuid.Parse(chi.URLParam(r, "empresaID"))

		if err != nil {
			m.error.NotFoundResponse(w, r, err)
			return
		}

		userID := util.GetUserFromContext(r).ID

		empresa, err := m.find.GetByID(r.Context(), userID, empresaID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				m.error.NotFoundResponse(w, r, err)
				return
			}

			m.error.InternalServerError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), util.EmpresaCtx, empresa)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
