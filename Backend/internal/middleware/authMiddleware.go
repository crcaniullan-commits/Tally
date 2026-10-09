package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/crcaniullan-commits/Tally/internal/auth"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserFinder interface {
	GetByID(context.Context, uuid.UUID) (*model.User, error)
}
type AuthMiddleware struct {
	error errorhandler.ErrorsResponse
	auth  auth.Authenticator
	find  UserFinder
}

func NewAuthMiddleware(e errorhandler.ErrorsResponse, a auth.Authenticator, f UserFinder) *AuthMiddleware {
	return &AuthMiddleware{e, a, f}
}

func (a *AuthMiddleware) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.error.UnauthorizedErrorResponse(w, r, fmt.Errorf("Authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			a.error.UnauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]

		jwtToken, err := a.auth.ValidateToken(token)

		if err != nil {
			a.error.UnauthorizedErrorResponse(w, r, err)
			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)

		if !ok {
			a.error.UnauthorizedErrorResponse(w, r, fmt.Errorf("invalid token claims"))
			return
		}

		sub, ok := claims["sub"].(string)

		if !ok {
			a.error.UnauthorizedErrorResponse(w, r, fmt.Errorf("sub claim missing or invalid"))
			return
		}

		userID, err := uuid.Parse(sub)

		if err != nil {
			a.error.UnauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := a.find.GetByID(ctx, userID)

		if err != nil {
			a.error.UnauthorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, util.UserCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (a *AuthMiddleware) CheckOwnership(requiredRole util.UserRole, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := util.GetUserFromContext(r)

		allowed := a.checkRole(requiredRole, user)

		if !allowed {
			a.error.ForbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func (a *AuthMiddleware) checkRole(requiredUser util.UserRole, user *model.User) bool {
	return user.Role == util.UserRole(requiredUser)
}
