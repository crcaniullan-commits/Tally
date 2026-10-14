package auth

import (
	"context"
	"errors"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthStore interface {
	Create(context.Context, *model.User) error
	GetByEmail(context.Context, string) (*model.User, error)
	setExpire(context.Context, time.Time, uuid.UUID) error
}

type Redeemer interface {
	Redeem(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error)
}

type AuthService struct {
	store         AuthStore
	redeemer      Redeemer
	authenticator Authenticator
	transactor    *dbtx.Transactor
}

func NewAuthService(store AuthStore, auth Authenticator, t *dbtx.Transactor, r Redeemer) *AuthService {
	return &AuthService{store, r, auth, t}
}

func (s *AuthService) RegisterUser(ctx context.Context, payload *CreateUserPayload) (string, error) {
	token, err := s.transactor.WithinTxWithResult(ctx, func(ctx context.Context) (string, error) {
		user := &model.User{
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

		accesskey, err := s.redeemer.Redeem(ctx, payload.Code, user.ID)

		if err != nil {
			return "", err
		}

		if err := s.store.setExpire(ctx, *accesskey.ExpiresAt, user.ID); err != nil {
			return "", err
		}

		return token, nil
	})

	if err != nil {
		return "", err
	}

	if token == "" {
		return "", errors.New("error al general el token")
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

func (s *AuthService) createToken(user *model.User) (string, error) {
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

func (s *AuthService) CreateMunicipal(ctx context.Context, payload *CreateMunicipalPayload) (string, error) {
	user := &model.User{
		Nombre: payload.Name,
		Email:  payload.Email,
		Role:   util.UserRoleMunicipal,
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
