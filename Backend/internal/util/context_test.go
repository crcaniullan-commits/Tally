package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var testUserID = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")

// newTestUser construye un usuario válido y reutilizable en los tests.
func newTestUser() *model.User {
	return &model.User{
		ID:     testUserID,
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   model.UserRoleUsuario,
		Rut:    model.RUT{Cuerpo: 19234567, DV: "K"},
	}
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("devuelve el usuario inyectado por el middleware", func(t *testing.T) {
		usuario := newTestUser()

		r := httptest.NewRequest(http.MethodPatch, "/users", nil)
		r = r.WithContext(context.WithValue(r.Context(), UserCtx, usuario))

		assert.Same(t, usuario, GetUserFromContext(r))
	})

	t.Run("devuelve nil si no hay usuario en el contexto", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/users", nil)

		assert.Nil(t, GetUserFromContext(r))
	})

	t.Run("devuelve nil si el valor del contexto es de otro tipo", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/users", nil)
		r = r.WithContext(context.WithValue(r.Context(), UserCtx, "no soy un usuario"))

		assert.Nil(t, GetUserFromContext(r))
	})
}
