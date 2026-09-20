package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreUser interface {
	Update(context.Context, *Users) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, string) (*Users, error)
	GetByID(context.Context, uuid.UUID) (*Users, error)
}

type UserService struct {
	store StoreUser
}

func NewUserService(store StoreUser) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID, payload *UpdateUserPayload) error {

	user, err := s.store.GetByID(ctx, userID)

	if err != nil {
		return err
	}

	if err = user.PasswordHash.Set(payload.Password); err != nil {
		return err
	}

	user.Nombre = payload.Name

	if err = s.store.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, userID uuid.UUID) error {

	if err := s.store.Delete(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetByRut(ctx context.Context, rut util.RUT) (*Users, error) {
	user, err := s.store.GetByRut(ctx, rut.String())

	if err != nil {
		return nil, err
	}

	return user, nil
}
