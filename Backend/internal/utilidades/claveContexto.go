package utilidades

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/model"
)

type claveUsuario string
type claveEmpresa string

const ContextoUsuario claveUsuario = "user_key"
const ContextoEmpresa claveEmpresa = "empresa_key"

// ObtenerUsuarioDelContexto devuelve el usuario que el middleware de autenticación
// inyectó en el contexto. Devuelve nil si no hay ninguno.
func ObtenerUsuarioDelContexto(r *http.Request) *model.User {
	user, _ := r.Context().Value(ContextoUsuario).(*model.User)
	return user
}

func ObtenerEmpresaDelContexto(r *http.Request) *model.Empresa {
	empresa, _ := r.Context().Value(ContextoEmpresa).(*model.Empresa)
	return empresa
}
