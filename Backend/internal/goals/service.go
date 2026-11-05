package goals

import (
	"context"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type GoalStorage interface {
	CreateGoal(context.Context, *model.Goal) error
	DeleteGoal(context.Context, uuid.UUID, uuid.UUID) error
	GetIncomesTotalByPeriod(context.Context, uuid.UUID, model.GoalPeriod) (int64, error)
	GetGoal(context.Context, uuid.UUID, uuid.UUID) (model.Goal, error)
	GetGoals(context.Context, uuid.UUID) ([]model.Goal, error)
}
type GoalService struct {
	store GoalStorage
}

func NewGoalService(store GoalStorage) *GoalService {
	return &GoalService{store: store}
}

func (s *GoalService) createGoal(ctx context.Context, payload CreateGoalPayload, empresaID uuid.UUID) (model.Goal, error) {

	goal := &model.Goal{
		Nombre:    payload.Nombre,
		Periodo:   payload.Periodo,
		MontoMeta: payload.Monto_meta,
		EmpresaID: empresaID,
	}

	if err := s.store.CreateGoal(ctx, goal); err != nil {
		return model.Goal{}, err
	}

	return *goal, nil
}

func (s *GoalService) DeleteGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.DeleteGoal(ctx, goalID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *GoalService) GetGoalState(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (GoalState, error) {

	goal, err := s.store.GetGoal(ctx, goalID, empresaID)

	if err != nil {
		return GoalState{}, err
	}

	if goal.MontoMeta <= 0 {
		return GoalState{}, fmt.Errorf("el goal %s tiene monto_meta inválido: %d", goal.ID, goal.MontoMeta)
	}

	monto, err := s.store.GetIncomesTotalByPeriod(ctx, empresaID, goal.Periodo)

	if err != nil {
		return GoalState{}, err
	}

	porActual := float64(monto) * 100 / float64(goal.MontoMeta)

	pendiente := goal.MontoMeta - monto
	if pendiente < 0 {
		pendiente = 0
	}

	porFaltante := 100 - porActual
	if porFaltante < 0 {
		porFaltante = 0
	}

	state := GoalState{
		MontoActual:        monto,
		PorcentajeActual:   porActual,
		MontoPendiente:     pendiente,
		PorcentajeFaltante: porFaltante,
	}

	return state, nil
}

func (s *GoalService) getGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {

	goal, err := s.store.GetGoal(ctx, goalID, empresaID)

	if err != nil {
		return model.Goal{}, err
	}

	return goal, nil
}

func (s *GoalService) getGoals(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {

	goals, err := s.store.GetGoals(ctx, empresaID)

	if err != nil {
		return nil, err
	}

	return goals, nil
}
