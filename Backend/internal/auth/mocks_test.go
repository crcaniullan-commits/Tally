package auth

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// AuthStoreMock mockea AuthStore (persistencia de auth).
type AuthStoreMock struct {
	mock.Mock
}

func (m *AuthStoreMock) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *AuthStoreMock) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

// setExpire implementa el metodo unexportado de AuthStore. Es posible porque el
// mock vive en el mismo package que la interfaz.
func (m *AuthStoreMock) setExpire(ctx context.Context, expiresAt time.Time, userID uuid.UUID) error {
	return m.Called(ctx, expiresAt, userID).Error(0)
}

// RedeemerMock mockea la interfaz Redeemer (resuelve el canje de una llave). En
// produccion la implementa el store de accessKeys.
type RedeemerMock struct {
	mock.Mock
}

func (m *RedeemerMock) Redeem(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error) {
	args := m.Called(ctx, code, userID)

	accessKey, _ := args.Get(0).(*model.AccessKey)

	return accessKey, args.Error(1)
}

// AuthenticatorMock mockea Authenticator para no firmar/validar JWT reales.
type AuthenticatorMock struct {
	mock.Mock
}

func (m *AuthenticatorMock) GenerateToken(claims jwt.Claims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *AuthenticatorMock) ValidateToken(token string) (*jwt.Token, error) {
	args := m.Called(token)

	t, _ := args.Get(0).(*jwt.Token)

	return t, args.Error(1)
}

// ServiceAuthMock mockea la capa de servicio (ServiceAuth) para probar
// AuthHandler de forma aislada.
type ServiceAuthMock struct {
	mock.Mock
}

func (m *ServiceAuthMock) RegisterUser(ctx context.Context, payload *CreateUserPayload) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *ServiceAuthMock) Login(ctx context.Context, payload *LoginUserPayload) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *ServiceAuthMock) CreateMunicipal(ctx context.Context, payload *CreateMunicipalPayload) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

var (
	testUserID        = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testMunicipalID   = uuid.MustParse("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f")
	testAccessKeyID   = uuid.MustParse("7a8b9c0d-1e2f-4a3b-8c9d-0e1f2a3b4c5d")
	testAccessKeyCode = "CODIGO-DE-PRUEBA"
	testPassword      = "ClaveSegura123"
	testToken         = "jwt.token.de.prueba"
)

// newTestUser construye un usuario valido con password ya hasheada.
func newTestUser() *model.User {
	user := &model.User{
		ID:     testUserID,
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   util.UserRoleUsuario,
		Rut:    util.RUT{Cuerpo: 19234567, DV: "K"},
	}

	_ = user.PasswordHash.Set(testPassword)

	return user
}

// newTestAccessKey construye una llave vigente y no canjeada.
func newTestAccessKey() model.AccessKey {
	expire := time.Date(2027, time.January, 15, 12, 0, 0, 0, time.UTC)

	return model.AccessKey{
		ID:        testAccessKeyID,
		Code:      testAccessKeyCode,
		CreatedBy: &testMunicipalID,
		ExpiresAt: &expire,
		CreatedAt: time.Now(),
	}
}

func newTestRegisterPayload() *CreateUserPayload {
	return &CreateUserPayload{
		Email:    "nuevo@correo.cl",
		Password: testPassword,
		Name:     "Nombre Nuevo",
		Rut:      "19.234.567-K",
		Code:     testAccessKeyCode,
	}
}

func newTestMunicipalPayload() *CreateMunicipalPayload {
	return &CreateMunicipalPayload{
		Email:    "municipal@correo.cl",
		Password: testPassword,
		Name:     "Nombre Municipal",
		Rut:      "19.234.567-K",
	}
}
