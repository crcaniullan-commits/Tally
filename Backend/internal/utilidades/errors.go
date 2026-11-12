package utilidades

import "errors"

var (
	ErrorNoEncontrado = errors.New("recurso no encontrado")
	ErrorConflicto    = errors.New("el recurso ya existe")
	ErrorNoAutorizado = errors.New("credenciales inválidas o inexistentes")
)
