package intermediario

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type BuscadorUsuario interface {
	ObtenerPorID(context.Context, uuid.UUID) (*model.User, error)
}
type MiddlewareAutenticacion struct {
	error errores.RespuestaErrores
	auth  auth.Authenticator
	find  BuscadorUsuario
}

func NuevoMiddlewareAutenticacion(e errores.RespuestaErrores, a auth.Authenticator, f BuscadorUsuario) *MiddlewareAutenticacion {
	return &MiddlewareAutenticacion{e, a, f}
}

func (a *MiddlewareAutenticacion) MiddlewareTokenAutenticacion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.error.ErrorNoAutorizado(w, r, fmt.Errorf("falta el encabezado Authorization"))
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			a.error.ErrorNoAutorizado(w, r, fmt.Errorf("el encabezado Authorization está mal formado"))
			return
		}

		token := parts[1]

		jwtToken, err := a.auth.ValidarToken(token)

		if err != nil {
			a.error.ErrorNoAutorizado(w, r, err)
			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)

		if !ok {
			a.error.ErrorNoAutorizado(w, r, fmt.Errorf("claims del token inválidos"))
			return
		}

		sub, ok := claims["sub"].(string)

		if !ok {
			a.error.ErrorNoAutorizado(w, r, fmt.Errorf("falta el claim sub o es inválido"))
			return
		}

		userID, err := uuid.Parse(sub)

		if err != nil {
			a.error.ErrorNoAutorizado(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := a.find.ObtenerPorID(ctx, userID)

		if err != nil {
			a.error.ErrorNoAutorizado(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, utilidades.ContextoUsuario, user)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (a *MiddlewareAutenticacion) VerificarRol(requiredRole utilidades.RolUsuarioTipo, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := utilidades.ObtenerUsuarioDelContexto(r)

		allowed := a.checkRole(requiredRole, user)

		if !allowed {
			a.error.ErrorProhibido(w, r)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func (a *MiddlewareAutenticacion) checkRole(requiredUser utilidades.RolUsuarioTipo, user *model.User) bool {
	return user.Role == utilidades.RolUsuarioTipo(requiredUser)
}
