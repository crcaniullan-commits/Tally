package auth

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/golang-jwt/jwt/v5"
)

type AuthStore interface {
	Create(context.Context, *users.Users) error
	GetByEmail(context.Context, string) (*users.Users, error)
}

type AuthService struct {
	store         AuthStore
	authenticator Authenticator
}

func NewAuthService(store AuthStore, auth Authenticator) *AuthService {
	return &AuthService{store, auth}
}

func (s *AuthService) RegisterUser(ctx context.Context, payload *CreateUserPayload) (string, error) {
	user := &users.Users{
		Nombre: payload.Name,
		Email:  payload.Email,
		Role:   util.UserRoleUsuario,
	}

	if err := user.PasswordHash.Set(payload.Password); err != nil {
		return "", err
	}

	rut, err := util.ParseRUT(payload.Rut)

	if err != nil {
		return "", err
	}

	user.Rut = rut

	err = s.store.Create(ctx, user)

	if err != nil {
		return "", err
	}

	token, err := s.createToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, payload *LoginUserPayload) (string, error) {
	user, err := s.store.GetByEmail(ctx, payload.Email)

	if err != nil {
		return "", err
	}

	if err = user.PasswordHash.Compare(payload.Password); err != nil {
		return "", err
	}

	token, err := s.createToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) createToken(user *users.Users) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(cExp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": cIss,
		"aud": cIss,
	}

	token, err := s.authenticator.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return token, err

}
