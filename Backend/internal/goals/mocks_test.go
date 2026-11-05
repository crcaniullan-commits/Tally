package goals

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// GoalStoreMock es un mock de GoalStorage para probar GoalService sin tocar la
// base de datos.
type GoalStoreMock struct {
	mock.Mock
}

func (m *GoalStoreMock) CreateGoal(ctx context.Context, goal *model.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *GoalStoreMock) DeleteGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, goalID, empresaID)
	return args.Error(0)
}

func (m *GoalStoreMock) GetIncomesTotalByPeriod(ctx context.Context, empresaID uuid.UUID, period model.GoalPeriod) (int64, error) {
	args := m.Called(ctx, empresaID, period)
	total, _ := args.Get(0).(int64)
	return total, args.Error(1)
}

func (m *GoalStoreMock) GetGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, goalID, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *GoalStoreMock) GetGoals(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {
	args := m.Called(ctx, empresaID)
	goals, _ := args.Get(0).([]model.Goal)
	return goals, args.Error(1)
}

// ServiceGoalsMock es un mock de ServiceGoals para probar GoalsHandler de forma
// aislada.
type ServiceGoalsMock struct {
	mock.Mock
}

func (m *ServiceGoalsMock) createGoal(ctx context.Context, payload CreateGoalPayload, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, payload, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *ServiceGoalsMock) DeleteGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, goalID, empresaID)
	return args.Error(0)
}

func (m *ServiceGoalsMock) GetGoalState(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (GoalState, error) {
	args := m.Called(ctx, goalID, empresaID)
	state, _ := args.Get(0).(GoalState)
	return state, args.Error(1)
}

func (m *ServiceGoalsMock) getGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {
	args := m.Called(ctx, goalID, empresaID)
	goal, _ := args.Get(0).(model.Goal)
	return goal, args.Error(1)
}

func (m *ServiceGoalsMock) getGoals(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {
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
func newTestPayloadGoal() CreateGoalPayload {
	return CreateGoalPayload{
		Nombre:     "Ahorrar",
		Periodo:    model.GoalPeriodMes,
		Monto_meta: 10000,
	}
}
