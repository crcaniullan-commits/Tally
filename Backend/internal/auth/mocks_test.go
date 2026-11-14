package auth

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// AuthStoreMock mockea RepositorioAuth (persistencia de auth).
type AuthStoreMock struct {
	mock.Mock
}

func (m *AuthStoreMock) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *AuthStoreMock) ObtenerPorEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

// establecerVencimiento implementa el metodo unexportado de RepositorioAuth. Es posible porque el
// mock vive en el mismo package que la interfaz.
func (m *AuthStoreMock) establecerVencimiento(ctx context.Context, expiresAt time.Time, userID uuid.UUID) error {
	return m.Called(ctx, expiresAt, userID).Error(0)
}

func (m *AuthStoreMock) CrearCodigoRecuperacion(ctx context.Context, codigo *model.CodigoRecuperacion) error {
	return m.Called(ctx, codigo).Error(0)
}

func (m *AuthStoreMock) ObtenerCodigoRecuperacion(ctx context.Context, userID uuid.UUID) (*model.CodigoRecuperacion, error) {
	args := m.Called(ctx, userID)

	codigo, _ := args.Get(0).(*model.CodigoRecuperacion)

	return codigo, args.Error(1)
}

func (m *AuthStoreMock) MarcarCodigoVerificado(ctx context.Context, codigoID uuid.UUID) error {
	return m.Called(ctx, codigoID).Error(0)
}

func (m *AuthStoreMock) MarcarCodigoUsado(ctx context.Context, codigoID uuid.UUID) error {
	return m.Called(ctx, codigoID).Error(0)
}

func (m *AuthStoreMock) ActualizarPassword(ctx context.Context, userID uuid.UUID, hash []byte) error {
	return m.Called(ctx, userID, hash).Error(0)
}

// RemitenteMock mockea la interfaz Remitente (envío del código de recuperación
// por correo) y guarda el último código enviado.
type RemitenteMock struct {
	mock.Mock

	UltimoCodigo  string
	UltimoDestino string
}

func (m *RemitenteMock) SendPasswordResetCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	m.UltimoCodigo = code
	m.UltimoDestino = to

	return m.Called(ctx, to, code, expiresAt).Error(0)
}

// CanjeadorMock mockea la interfaz Canjeador (resuelve el canje de una llave). En
// produccion la implementa el store de accessKeys.
type CanjeadorMock struct {
	mock.Mock
}

func (m *CanjeadorMock) Canjear(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error) {
	args := m.Called(ctx, code, userID)

	accessKey, _ := args.Get(0).(*model.AccessKey)

	return accessKey, args.Error(1)
}

// AutenticadorMock mockea Authenticator para no firmar/validar JWT reales.
type AutenticadorMock struct {
	mock.Mock
}

func (m *AutenticadorMock) GenerarToken(claims jwt.Claims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *AutenticadorMock) ValidarToken(token string) (*jwt.Token, error) {
	args := m.Called(token)

	t, _ := args.Get(0).(*jwt.Token)

	return t, args.Error(1)
}

// IServiceAuthMock mockea la capa de servicio (IServiceAuth) para probar
// HandlerAuth de forma aislada.
type IServiceAuthMock struct {
	mock.Mock
}

func (m *IServiceAuthMock) RegistrarUsuario(ctx context.Context, payload *PayloadCrearUsuario) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *IServiceAuthMock) Login(ctx context.Context, payload *PayloadInicioSesion) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *IServiceAuthMock) CrearMunicipal(ctx context.Context, payload *PayloadCrearMunicipal) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *IServiceAuthMock) SolicitarRecuperacion(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

func (m *IServiceAuthMock) VerificarCodigoRecuperacion(ctx context.Context, email, codigo string) error {
	return m.Called(ctx, email, codigo).Error(0)
}

func (m *IServiceAuthMock) ConfirmarRecuperacion(ctx context.Context, email, codigo, password string) error {
	return m.Called(ctx, email, codigo, password).Error(0)
}

var (
	testUserID               = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testMunicipalID          = uuid.MustParse("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f")
	testAccessKeyID          = uuid.MustParse("7a8b9c0d-1e2f-4a3b-8c9d-0e1f2a3b4c5d")
	testAccessKeyCode        = "CODIGO-DE-PRUEBA"
	testPassword             = "ClaveSegura123"
	testToken                = "jwt.token.de.prueba"
	testCodigoRecuperacion   = "123456"
	testCodigoRecuperacionID = uuid.MustParse("2b3c4d5e-6f70-4812-93a4-b5c6d7e8f901")
)

// newTestUser construye un usuario valido con password ya hasheada.
func newTestUser() *model.User {
	user := &model.User{
		ID:     testUserID,
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   utilidades.RolUsuario,
		Rut:    utilidades.RUT{Cuerpo: 19234567, DV: "K"},
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

func newTestRegisterPayload() *PayloadCrearUsuario {
	return &PayloadCrearUsuario{
		Email:    "nuevo@correo.cl",
		Password: testPassword,
		Nombre:   "Nombre Nuevo",
		Rut:      "19.234.567-K",
		Code:     testAccessKeyCode,
	}
}

func newTestMunicipalPayload() *PayloadCrearMunicipal {
	return &PayloadCrearMunicipal{
		Email:    "municipal@correo.cl",
		Password: testPassword,
		Nombre:   "Nombre Municipal",
		Rut:      "19.234.567-K",
	}
}

// newTestCodigoRecuperacion arma un código de recuperación vigente, sin
// verificar y sin usar.
func newTestCodigoRecuperacion() *model.CodigoRecuperacion {
	return &model.CodigoRecuperacion{
		ID:        testCodigoRecuperacionID,
		UserID:    testUserID,
		Codigo:    hashearCodigo(testCodigoRecuperacion),
		ExpiraEn:  time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
}
