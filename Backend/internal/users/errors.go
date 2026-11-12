package users

import (
	"errors"
)

var (
	ErrDuplicateEmail    = errors.New("ya existe un usuario con ese correo")
	ErrDuplicateUsername = errors.New("ya existe un usuario con ese nombre de usuario")
	ErrDuplicateRut      = errors.New("ya existe un usuario con ese RUT")
)
