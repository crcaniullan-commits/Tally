package users

import (
	"errors"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
	ErrDuplicateRut      = errors.New("a user with that rut already exists")
	ErrCodeRedemed       = errors.New("the code was already exchanged")
	ErrCodeVencido       = errors.New("the code is expired")
	ErrCodeRevocado      = errors.New("the code was revoked")
)
