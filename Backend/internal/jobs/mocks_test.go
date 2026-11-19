package jobs

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type JobsStoreMock struct {
	mock.Mock
}

func (m *JobsStoreMock) UsuariosSinMovimientos(ctx context.Context, start, end time.Time) ([]uuid.UUID, error) {
	args := m.Called(ctx, start, end)
	res, _ := args.Get(0).([]uuid.UUID)
	return res, args.Error(1)
}

func (m *JobsStoreMock) CobrosVencidosPorUsuario(ctx context.Context, day string) ([]ResumenCobrosVencidos, error) {
	args := m.Called(ctx, day)
	res, _ := args.Get(0).([]ResumenCobrosVencidos)
	return res, args.Error(1)
}

type NotificadorMock struct {
	mock.Mock
}

func (m *NotificadorMock) Send(ctx context.Context, userID uuid.UUID, n model.Notification) error {
	args := m.Called(ctx, userID, n)
	return args.Error(0)
}

func (m *NotificadorMock) Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error) {
	args := m.Called(ctx, userID, kind, day)
	return args.Bool(0), args.Error(1)
}

func (m *NotificadorMock) Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error {
	args := m.Called(ctx, userID, kind, day)
	return args.Error(0)
}

type IServicioJobsMock struct {
	mock.Mock
}

func (m *IServicioJobsMock) RecordatorioDiario(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *IServicioJobsMock) CobrosVencidos(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *IServicioJobsMock) EjecutarTodos(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

var (
	testUserID = uuid.New()
)
