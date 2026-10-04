package users

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreUser interface {
	Update(context.Context, *model.User) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, string) (*model.User, error)
	GetExpire(context.Context, string) (model.AccessKey, error)
	setExpire(context.Context, time.Time, uuid.UUID) error
}

type UserService struct {
	store StoreUser
}

func NewUserService(store StoreUser) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Update(ctx context.Context, user *model.User, payload *UpdateUserPayload) error {
	if payload.Password != "" {
		if err := user.PasswordHash.Set(payload.Password); err != nil {
			return err
		}
	}

	if payload.Name != "" {
		user.Nombre = payload.Name
	}

	if err := s.store.Update(ctx, user); err != nil {
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

func (s *UserService) GetByRut(ctx context.Context, rut util.RUT) (*model.User, error) {
	user, err := s.store.GetByRut(ctx, rut.String())

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ExchangeCode(ctx context.Context, code string, userID uuid.UUID) error {
	accessKey, err := s.store.GetExpire(ctx, code)

	if err != nil {
		return err
	}

	if accessKey.RedeemedAt != nil {
		return ErrCodeRedemed
	}

	if accessKey.RevokedAt != nil {
		return ErrCodeRevocado
	}

	// expires_at es nullable: una llave sin vencimiento no se puede canjear, y
	// desreferenciarla a ciegas haría entrar el service en panic.
	if accessKey.ExpiresAt == nil {
		return ErrCodeVencido
	}

	if !accessKey.ExpiresAt.After(time.Now()) {
		return ErrCodeVencido
	}

	if err := s.store.setExpire(ctx, *accessKey.ExpiresAt, userID); err != nil {
		return err
	}

	return nil
}
