package goals

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestGoalService_createGoal(t *testing.T) {
	t.Run("crea el goal con el payload y la empresa", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		payload := newTestPayloadGoal()

		store.On("CreateGoal", mock.Anything, mock.MatchedBy(func(goal *model.Goal) bool {
			return goal.Nombre == payload.Nombre &&
				goal.Periodo == payload.Periodo &&
				goal.MontoMeta == payload.Monto_meta &&
				goal.EmpresaID == testEmpresaID
		})).Return(nil).Run(func(args mock.Arguments) {
			goal := args.Get(1).(*model.Goal)
			goal.ID = testGoalID
			goal.CreatedAt = testGoalCreatedAt
		}).Once()

		goal, err := service.createGoal(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testGoalID, goal.ID)
		assert.Equal(t, testEmpresaID, goal.EmpresaID)
		assert.Equal(t, payload.Nombre, goal.Nombre)
		assert.Equal(t, payload.Periodo, goal.Periodo)
		assert.Equal(t, payload.Monto_meta, goal.MontoMeta)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("CreateGoal", mock.Anything, mock.AnythingOfType("*model.Goal")).
			Return(errStoreBoom).Once()

		goal, err := service.createGoal(context.Background(), newTestPayloadGoal(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, model.Goal{}, goal)
		store.AssertExpectations(t)
	})
}

func TestGoalService_DeleteGoal(t *testing.T) {
	t.Run("elimina el goal", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("DeleteGoal", mock.Anything, testGoalID, testEmpresaID).Return(nil).Once()

		err := service.DeleteGoal(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("DeleteGoal", mock.Anything, testGoalID, testEmpresaID).
			Return(errStoreBoom).Once()

		err := service.DeleteGoal(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestGoalService_GetGoalState(t *testing.T) {
	t.Run("sin ingresos deja la meta completa pendiente", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()
		store.On("GetIncomesTotalByPeriod", mock.Anything, testEmpresaID, goal.Periodo).
			Return(int64(0), nil).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, GoalState{
			MontoActual:        0,
			PorcentajeActual:   0,
			MontoPendiente:     goal.MontoMeta,
			PorcentajeFaltante: 100,
		}, state)
		store.AssertExpectations(t)
	})

	t.Run("calcula el avance parcial", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()
		store.On("GetIncomesTotalByPeriod", mock.Anything, testEmpresaID, goal.Periodo).
			Return(int64(2500), nil).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, int64(2500), state.MontoActual)
		assert.Equal(t, float64(25), state.PorcentajeActual)
		assert.Equal(t, int64(7500), state.MontoPendiente)
		assert.Equal(t, float64(75), state.PorcentajeFaltante)
		store.AssertExpectations(t)
	})

	t.Run("meta cumplida al 100 por ciento", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()
		store.On("GetIncomesTotalByPeriod", mock.Anything, testEmpresaID, goal.Periodo).
			Return(goal.MontoMeta, nil).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, float64(100), state.PorcentajeActual)
		assert.Equal(t, int64(0), state.MontoPendiente)
		assert.Equal(t, float64(0), state.PorcentajeFaltante)
		store.AssertExpectations(t)
	})

	t.Run("meta superada no deja pendiente negativo", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()
		store.On("GetIncomesTotalByPeriod", mock.Anything, testEmpresaID, goal.Periodo).
			Return(int64(15000), nil).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, float64(150), state.PorcentajeActual)
		assert.Equal(t, int64(0), state.MontoPendiente)
		assert.Equal(t, float64(0), state.PorcentajeFaltante)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error al buscar el goal", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).
			Return(model.Goal{}, errStoreBoom).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, GoalState{}, state)
		store.AssertNotCalled(t, "GetIncomesTotalByPeriod", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error al sumar los ingresos", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()
		store.On("GetIncomesTotalByPeriod", mock.Anything, testEmpresaID, goal.Periodo).
			Return(int64(0), errStoreBoom).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, GoalState{}, state)
		store.AssertExpectations(t)
	})

	t.Run("rechaza un goal con monto_meta invalido", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()
		goal.MontoMeta = 0

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()

		state, err := service.GetGoalState(context.Background(), testGoalID, testEmpresaID)

		require.Error(t, err)
		assert.Equal(t, GoalState{}, state)
		store.AssertNotCalled(t, "GetIncomesTotalByPeriod", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})
}

func TestGoalService_getGoal(t *testing.T) {
	t.Run("devuelve el goal", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goal := newTestGoal()

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).Return(goal, nil).Once()

		result, err := service.getGoal(context.Background(), testGoalID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, goal, result)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("GetGoal", mock.Anything, testGoalID, testEmpresaID).
			Return(model.Goal{}, errStoreBoom).Once()

		result, err := service.getGoal(context.Background(), testGoalID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, model.Goal{}, result)
		store.AssertExpectations(t)
	})
}

func TestGoalService_getGoals(t *testing.T) {
	t.Run("devuelve los goals de la empresa", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)
		goals := []model.Goal{newTestGoal()}

		store.On("GetGoals", mock.Anything, testEmpresaID).Return(goals, nil).Once()

		result, err := service.getGoals(context.Background(), testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, goals, result)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(GoalStoreMock)
		service := NewGoalService(store)

		store.On("GetGoals", mock.Anything, testEmpresaID).
			Return(nil, errStoreBoom).Once()

		result, err := service.getGoals(context.Background(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, result)
		store.AssertExpectations(t)
	})
}
