package utilidades

import (
	"net/http"
)

type MiddlewareAutenticacion interface {
	VerificarRol(RolUsuarioTipo, http.HandlerFunc) http.HandlerFunc
}
