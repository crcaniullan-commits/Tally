package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// StoreUserMock es un mock de la capa de persistencia (StoreUser) usado para
// probar UserService sin tocar la base de datos.
type StoreUserMock struct {
	mock.Mock
}

func (m *StoreUserMock) Update(ctx context.Context, user *Users) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *StoreUserMock) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *StoreUserMock) GetByRut(ctx context.Context, rut string) (*Users, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*Users)

	return user, args.Error(1)
}

// ServiceUsersMock es un mock de la capa de servicio (ServiceUsers) usado para
// probar UsersHandler de forma aislada.
type ServiceUsersMock struct {
	mock.Mock
}

func (m *ServiceUsersMock) Update(ctx context.Context, user *Users, payload *UpdateUserPayload) error {
	args := m.Called(ctx, user, payload)
	return args.Error(0)
}

func (m *ServiceUsersMock) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *ServiceUsersMock) GetByRut(ctx context.Context, rut util.RUT) (*Users, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*Users)

	return user, args.Error(1)
}

// newTestUser construye un usuario válido y reutilizable en los tests.
func newTestUser() *Users {
	return &Users{
		ID:     uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b"),
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   util.UserRoleUsuario,
		Rut:    util.RUT{Cuerpo: 19234567, DV: "K"},
	}
}
