package users

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreUser interface {
	Update(context.Context, *model.User) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, string) (*model.User, error)
	setExpire(context.Context, time.Time, uuid.UUID) error
}

type Redeemer interface {
	Redeem(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error)
}

type UserService struct {
	store      StoreUser
	redeemer   Redeemer
	transactor *dbtx.Transactor
}

func NewUserService(store StoreUser, redeemer Redeemer, transactor *dbtx.Transactor) *UserService {
	return &UserService{store: store, redeemer: redeemer, transactor: transactor}
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
	return s.transactor.WithinTx(ctx, func(ctx context.Context) error {
		accesskey, err := s.redeemer.Redeem(ctx, code, userID)

		if err != nil {
			return err
		}

		return s.store.setExpire(ctx, *accesskey.ExpiresAt, userID)
	})
}
