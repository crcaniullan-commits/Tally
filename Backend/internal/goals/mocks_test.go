package goals

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// GoalStoreMock es un mock de RepositorioMetas para probar ServicioMetas sin tocar la
// base de datos.
type GoalStoreMock struct {
	mock.Mock
}

func (m *GoalStoreMock) CrearMeta(ctx context.Context, goal *model.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *GoalStoreMock) EliminarMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, metaID, empresaID)
	return args.Error(0)
}

func (m *GoalStoreMock) ObtenerTotalIngresosPorPeriodo(ctx context.Context, empresaID uuid.UUID, period model.GoalPeriod) (int64, error) {
	args := m.Called(ctx, empresaID, period)
	total, _ := args.Get(0).(int64)
	return total, args.Error(1)
}

func (m *GoalStoreMock) ObtenerMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, metaID, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *GoalStoreMock) ObtenerMetas(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {
	args := m.Called(ctx, empresaID)
	goals, _ := args.Get(0).([]model.Goal)
	return goals, args.Error(1)
}

// IServiceGoalsMock es un mock de IServiceMetas para probar HandlerMetas de forma
// aislada.
type IServiceGoalsMock struct {
	mock.Mock
}

func (m *IServiceGoalsMock) CrearMeta(ctx context.Context, payload PayloadCrearMeta, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, payload, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *IServiceGoalsMock) EliminarMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, metaID, empresaID)
	return args.Error(0)
}

func (m *IServiceGoalsMock) ObtenerEstadoMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) (EstadoMeta, error) {
	args := m.Called(ctx, metaID, empresaID)
	state, _ := args.Get(0).(EstadoMeta)
	return state, args.Error(1)
}

func (m *IServiceGoalsMock) ObtenerMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, metaID, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *IServiceGoalsMock) ObtenerMetas(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {
	args := m.Called(ctx, empresaID)
	goals, _ := args.Get(0).([]model.Goal)
	return goals, args.Error(1)
}

var (
	testUserID    = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testGoalID    = uuid.MustParse("d4e5f6a7-8b9c-4d0e-9f1a-2b3c4d5e6f7a")
)

var testGoalCreatedAt = time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)

// newTestGoal construye un model.Goal válido y reutilizable en los tests.
func newTestGoal() model.Goal {
	return model.Goal{
		ID:        testGoalID,
		EmpresaID: testEmpresaID,
		Nombre:    "Ahorrar",
		Periodo:   model.GoalPeriodMes,
		MontoMeta: 10000,
		CreatedAt: testGoalCreatedAt,
	}
}

// newTestPayloadGoal arma el payload de alta con valores que pasan la validación.
func newTestPayloadGoal() PayloadCrearMeta {
	return PayloadCrearMeta{
		Nombre:    "Ahorrar",
		Periodo:   model.GoalPeriodMes,
		MontoMeta: 10000,
	}
}
