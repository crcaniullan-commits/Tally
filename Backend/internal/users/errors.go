package users

import (
	"errors"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
	ErrDuplicateRut      = errors.New("a user with that rut already exists")
)
