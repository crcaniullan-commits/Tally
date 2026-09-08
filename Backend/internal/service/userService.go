package service

import (
	"context"
	"net/http"
	"uuid"

	"github.com/crcaniullan-commits/Tally/cmd/src/httputil"
	"github.com/crcaniullan-commits/Tally/internal/store"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/go-chi/chi/v5"
)

type UserStore interface {
	Create(context.Context, *store.Users) (*store.Users, error)
	Update(context.Context, *store.Users) error
	Delete(context.Context, uuid.UUID) error
	GetByRut(context.Context, string) (*store.Users, error)
	GetByEmail(context.Context, string) (*store.Users, error)
	GetByID(context.Context, uuid.UUID) (*store.Users, error)
}

type CreateUserPaylaod struct {
	Email    string `json:"email" validate:"required,max=100,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
	Rut      string `json:"rut" validate:"required,max=10"`
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

	rut, err := util.ParseRUT(payload.Rut)

	if err != nil {
		return nil, err
	}

	if err = rut.Validar(); err != nil {
		return nil, err
	}

	user := &store.Users{
		Email:        payload.Email,
		PasswordHash: payload.Password,
		Nombre:       payload.Name,
		Role:         util.UserRoleUsuario,
		Rut:          rut,
	}

	ctx := r.Context()

	if err = s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

type UpdateUserPayload struct {
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"nombre" validate:"required,max=100"`
}

func (s *UserService) Update(w http.ResponseWriter, r *http.Request) error {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		return err
	}

	var payload UpdateUserPayload
	if err = httputil.ReadJSON(w, r, &payload); err != nil {
		return err
	}

	if err = httputil.Validate.Struct(payload); err != nil {
		return err
	}

	user, err := s.store.GetByID(r.Context(), userID)

	if err != nil {
		return err
	}

	user.PasswordHash = payload.Password
	user.Nombre = payload.Name

	if err = s.store.Update(r.Context(), user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Delete(w http.ResponseWriter, r *http.Request) error {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		return err
	}

	if err = s.store.Delete(r.Context(), userID); err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetByRut(w http.ResponseWriter, r *http.Request) (*store.Users, error) {
	rut, err := util.ParseRUT(chi.URLParam(r, "rut"))

	if err != nil {
		return nil, err
	}

	user, err := s.store.GetByRut(r.Context(), rut.String())

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByEmail(w http.ResponseWriter, r *http.Request) (*store.Users, error) {
	email := chi.URLParam(r, "email")

	user, err := s.store.GetByEmail(r.Context(), email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
