package util

import (
	"net/http"
)

type MiddlewareAuth interface {
	CheckOwnership(UserRole, http.HandlerFunc) http.HandlerFunc
}
