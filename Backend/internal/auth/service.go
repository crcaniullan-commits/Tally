package auth

import (
	"context"
	"errors"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RepositorioAuth interface {
	Create(context.Context, *model.User) error
	ObtenerPorEmail(context.Context, string) (*model.User, error)
	establecerVencimiento(context.Context, time.Time, uuid.UUID) error
}

type Canjeador interface {
	Canjear(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error)
}

type ServicioAuth struct {
	store         RepositorioAuth
	redeemer      Canjeador
	authenticator Authenticator
	transactor    *bdtx.Transactor
}

func NuevoServicioAuth(store RepositorioAuth, auth Authenticator, t *bdtx.Transactor, r Canjeador) *ServicioAuth {
	return &ServicioAuth{store, r, auth, t}
}

func (s *ServicioAuth) RegistrarUsuario(ctx context.Context, payload *PayloadCrearUsuario) (string, error) {
	token, err := s.transactor.WithinTxWithResult(ctx, func(ctx context.Context) (string, error) {
		user := &model.User{
			Nombre: payload.Nombre,
			Email:  payload.Email,
			Role:   utilidades.RolUsuario,
		}

		if err := user.PasswordHash.Set(payload.Password); err != nil {
			return "", err
		}

		rut, err := utilidades.ParseRUT(payload.Rut)

		if err != nil {
			return "", err
		}

		user.Rut = rut

		err = s.store.Create(ctx, user)

		if err != nil {
			return "", err
		}

		token, err := s.crearToken(user)

		if err != nil {
			return "", err
		}

		accesskey, err := s.redeemer.Canjear(ctx, payload.Code, user.ID)

		if err != nil {
			return "", err
		}

		if err := s.store.establecerVencimiento(ctx, *accesskey.ExpiresAt, user.ID); err != nil {
			return "", err
		}

		return token, nil
	})

	if err != nil {
		return "", err
	}

	if token == "" {
		return "", errors.New("error al generar el token")
	}

	return token, nil

}

func (s *ServicioAuth) Login(ctx context.Context, payload *PayloadInicioSesion) (string, error) {
	user, err := s.store.ObtenerPorEmail(ctx, payload.Email)

	if err != nil {
		return "", err
	}

	if err = user.PasswordHash.Compare(payload.Password); err != nil {
		return "", err
	}

	token, err := s.crearToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *ServicioAuth) crearToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(cExp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": cIss,
		"aud": cIss,
	}

	token, err := s.authenticator.GenerarToken(claims)
	if err != nil {
		return "", err
	}

	return token, err

}

func (s *ServicioAuth) CrearMunicipal(ctx context.Context, payload *PayloadCrearMunicipal) (string, error) {
	user := &model.User{
		Nombre: payload.Nombre,
		Email:  payload.Email,
		Role:   utilidades.RolMunicipal,
	}

	if err := user.PasswordHash.Set(payload.Password); err != nil {
		return "", err
	}

	rut, err := utilidades.ParseRUT(payload.Rut)

	if err != nil {
		return "", err
	}

	user.Rut = rut

	err = s.store.Create(ctx, user)

	if err != nil {
		return "", err
	}

	token, err := s.crearToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}
