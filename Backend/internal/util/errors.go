package util

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resources already exists")
	ErrUnauthorized = errors.New("Credenciales invalidas o inexistentes")
)
