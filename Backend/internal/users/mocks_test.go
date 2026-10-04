package users

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// StoreUserMock es un mock de la capa de persistencia (StoreUser) usado para
// probar UserService sin tocar la base de datos.
type StoreUserMock struct {
	mock.Mock
}

func (m *StoreUserMock) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *StoreUserMock) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *StoreUserMock) GetByRut(ctx context.Context, rut string) (*model.User, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

func (m *StoreUserMock) GetExpire(ctx context.Context, code string) (model.AccessKey, error) {
	args := m.Called(ctx, code)

	accessKey, _ := args.Get(0).(model.AccessKey)

	return accessKey, args.Error(1)
}

// setExpire implementa el metodo unexportado de StoreUser. Es posible porque el
// mock vive en el mismo package que la interfaz.
func (m *StoreUserMock) setExpire(ctx context.Context, expiresAt time.Time, userID uuid.UUID) error {
	args := m.Called(ctx, expiresAt, userID)
	return args.Error(0)
}

// ServiceUsersMock es un mock de la capa de servicio (ServiceUsers) usado para
// probar UsersHandler de forma aislada.
type ServiceUsersMock struct {
	mock.Mock
}

func (m *ServiceUsersMock) Update(ctx context.Context, user *model.User, payload *UpdateUserPayload) error {
	args := m.Called(ctx, user, payload)
	return args.Error(0)
}

func (m *ServiceUsersMock) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *ServiceUsersMock) GetByRut(ctx context.Context, rut util.RUT) (*model.User, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

func (m *ServiceUsersMock) ExchangeCode(ctx context.Context, code string, userID uuid.UUID) error {
	args := m.Called(ctx, code, userID)
	return args.Error(0)
}

var testUserID = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")

// newTestUser construye un usuario válido y reutilizable en los tests.
func newTestUser() *model.User {
	return &model.User{
		ID:     testUserID,
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   util.UserRoleUsuario,
		Rut:    util.RUT{Cuerpo: 19234567, DV: "K"},
	}
}

var (
	testMunicipalID  = uuid.MustParse("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f")
	testAccessKeyID  = uuid.MustParse("7a8b9c0d-1e2f-4a3b-8c9d-0e1f2a3b4c5d")
	testAccessKeyCod = "CODIGO-DE-PRUEBA"
)

// newTestAccessKey construye una llave de acceso vigente, no canjeada y no
// revocada: el estado en el que un codigo se puede canjear. Los tests que
// necesitan otro estado parten de aca y cambian el campo correspondiente.
func newTestAccessKey() model.AccessKey {
	expire := time.Now().Add(24 * time.Hour)

	return model.AccessKey{
		ID:        testAccessKeyID,
		Code:      testAccessKeyCod,
		CreatedBy: testMunicipalID,
		ExpiresAt: &expire,
		CreatedAt: time.Now(),
	}
}
