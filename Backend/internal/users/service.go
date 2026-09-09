package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type UserStore interface {
	Create(context.Context, *Users) (*Users, error)
	Update(context.Context, *Users) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, string) (*Users, error)
	GetByEmail(context.Context, string) (*Users, error)
	GetByID(context.Context, uuid.UUID) (*Users, error)
}

type UserService struct {
	store StoreUser
}

func NewUserService(store StoreUser) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Create(ctx context.Context, payload *CreateUserPaylaod) (*Users, error) {
	rut, err := util.ParseRUT(payload.Rut)

	if err != nil {
		return nil, err
	}

	if err = rut.Validar(); err != nil {
		return nil, err
	}

	user := &Users{
		Email:        payload.Email,
		PasswordHash: payload.Password,
		Nombre:       payload.Name,
		Role:         util.UserRoleUsuario,
		Rut:          rut,
	}

	if err = s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID, payload *UpdateUserPayload) error {

	user, err := s.store.GetByID(ctx, userID)

	if err != nil {
		return err
	}

	user.PasswordHash = payload.Password
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

func (s *UserService) GetByEmail(ctx context.Context, email string) (*Users, error) {
	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
