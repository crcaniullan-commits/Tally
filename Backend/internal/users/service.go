package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioUsuarios interface {
	Actualizar(context.Context, *model.User) error
	Eliminar(context.Context, uuid.UUID) error
	ObtenerPorRUT(context.Context, string) (*model.User, error)
}

type ServicioUsuarios struct {
	store RepositorioUsuarios
}

func NuevoServicioUsuarios(store RepositorioUsuarios) *ServicioUsuarios {
	return &ServicioUsuarios{store: store}
}

func (s *ServicioUsuarios) Actualizar(ctx context.Context, user *model.User, payload *PayloadActualizarUsuario) error {
	if payload.Password != "" {
		if err := user.PasswordHash.Set(payload.Password); err != nil {
			return err
		}
	}

	if payload.Nombre != "" {
		user.Nombre = payload.Nombre
	}

	if err := s.store.Actualizar(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *ServicioUsuarios) Eliminar(ctx context.Context, userID uuid.UUID) error {
	if err := s.store.Eliminar(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (s *ServicioUsuarios) ObtenerPorRUT(ctx context.Context, rut utilidades.RUT) (*model.User, error) {
	user, err := s.store.ObtenerPorRUT(ctx, rut.String())

	if err != nil {
		return nil, err
	}

	return user, nil
}
