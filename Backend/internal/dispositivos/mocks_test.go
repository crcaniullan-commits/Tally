package dispositivos

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type DispositivosStoreMock struct {
	mock.Mock
}

func (m *DispositivosStoreMock) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
	args := m.Called(ctx, userID, token, platform)
	return args.Error(0)
}

func (m *DispositivosStoreMock) Eliminar(ctx context.Context, userID uuid.UUID, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *DispositivosStoreMock) ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	tokens, _ := args.Get(0).([]string)
	return tokens, args.Error(1)
}

func (m *DispositivosStoreMock) ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	args := m.Called(ctx, userID)
	dispositivos, _ := args.Get(0).([]model.DeviceToken)
	return dispositivos, args.Error(1)
}

func (m *DispositivosStoreMock) EliminarTokens(ctx context.Context, tokens []string) error {
	args := m.Called(ctx, tokens)
	return args.Error(0)
}

type IServiceDispositivosMock struct {
	mock.Mock
}

func (m *IServiceDispositivosMock) Registrar(ctx context.Context, userID uuid.UUID, payload PayloadDispositivo) error {
	args := m.Called(ctx, userID, payload)
	return args.Error(0)
}

func (m *IServiceDispositivosMock) Desvincular(ctx context.Context, userID uuid.UUID, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *IServiceDispositivosMock) Listar(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	args := m.Called(ctx, userID)
	dispositivos, _ := args.Get(0).([]model.DeviceToken)
	return dispositivos, args.Error(1)
}

var (
	testUserID   = uuid.New()
	testDeviceID = uuid.New()
	testToken    = "ExponentPushToken[xxxxxxxxxxxxxxxxxxxxxx]"
	testNow      = time.Now().UTC()
)
