package service

import (
	"context"
	"net/http"
	"uuid"

	"github.com/crcaniullan-commits/Tally/cmd/src/httputil"
	"github.com/crcaniullan-commits/Tally/internal/store"
	"github.com/crcaniullan-commits/Tally/internal/util"
)

type UserStore interface {
	Create(http.ResponseWriter, *http.Request) (*store.Users, error)
	Update(context.Context, uuid.UUID) error
	Delete(context.Context, uuid.UUID) error
	GetByID(context.Context, uuid.UUID) (*store.Users, error)
}

type CreateUserPaylaod struct {
	Email    string `json:"email" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
}

type UserService struct {
	store store.StoreUser
}

func NewUserService(store store.StoreUser) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Create(w http.ResponseWriter, r *http.Request) (*store.Users, error) {
	var payload CreateUserPaylaod
	if err := httputil.ReadJSON(w, r, &payload); err != nil {
		return nil, err
	}

	if err := httputil.Validate.Struct(payload); err != nil {
		return nil, err
	}

	user := &store.Users{
		Email:        payload.Email,
		PasswordHash: payload.Password,
		Nombre:       payload.Name,
		Role:         util.UserRoleUsuario,
	}

	ctx := r.Context()

	if err := s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *UserService) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*store.Users, error) {
	return nil, nil
}
