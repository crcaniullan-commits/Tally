package notificaciones

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type NotificacionesStoreMock struct {
	mock.Mock
}

func (m *NotificacionesStoreMock) Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error) {
	args := m.Called(ctx, userID, kind, day)
	return args.Bool(0), args.Error(1)
}

func (m *NotificacionesStoreMock) Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error {
	args := m.Called(ctx, userID, kind, day)
	return args.Error(0)
}

func (m *NotificacionesStoreMock) ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error) {
	args := m.Called(ctx, userID)
	logs, _ := args.Get(0).([]model.NotificationLog)
	return logs, args.Error(1)
}

type BuscadorDispositivosMock struct {
	mock.Mock
}

func (m *BuscadorDispositivosMock) ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	tokens, _ := args.Get(0).([]string)
	return tokens, args.Error(1)
}

func (m *BuscadorDispositivosMock) EliminarTokensInvalidos(ctx context.Context, tokens []string) error {
	args := m.Called(ctx, tokens)
	return args.Error(0)
}

type EnviadorPushMock struct {
	mock.Mock
}

func (m *EnviadorPushMock) EnviarPush(ctx context.Context, tokens []string, n model.Notification) ([]string, error) {
	args := m.Called(ctx, tokens, n)
	invalidos, _ := args.Get(0).([]string)
	return invalidos, args.Error(1)
}

type IServiceNotificacionesMock struct {
	mock.Mock
}

func (m *IServiceNotificacionesMock) Send(ctx context.Context, userID uuid.UUID, n model.Notification) error {
	args := m.Called(ctx, userID, n)
	return args.Error(0)
}

func (m *IServiceNotificacionesMock) Listar(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error) {
	args := m.Called(ctx, userID)
	logs, _ := args.Get(0).([]model.NotificationLog)
	return logs, args.Error(1)
}

var (
	testUserID = uuid.New()
	testNow    = time.Now().UTC()
	testTokens = []string{"ExponentPushToken[1111111111111111111111]"}
)
