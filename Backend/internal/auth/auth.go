package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	GenerarToken(claims jwt.Claims) (string, error)
	ValidarToken(token string) (*jwt.Token, error)
}
