package util

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/model"
)

type userKey string

const UserCtx userKey = "user_key"

// GetUserFromContext devuelve el usuario que el middleware de autenticación
// inyectó en el contexto. Devuelve nil si no hay ninguno.
func GetUserFromContext(r *http.Request) *model.User {
	user, _ := r.Context().Value(UserCtx).(*model.User)
	return user
}
