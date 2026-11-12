package goals

import (
	"context"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type RepositorioMetas interface {
	CrearMeta(context.Context, *model.Goal) error
	EliminarMeta(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerTotalIngresosPorPeriodo(context.Context, uuid.UUID, model.GoalPeriod) (int64, error)
	ObtenerMeta(context.Context, uuid.UUID, uuid.UUID) (model.Goal, error)
	ObtenerMetas(context.Context, uuid.UUID) ([]model.Goal, error)
}
type ServicioMetas struct {
	store RepositorioMetas
}

func NuevoServicioMetas(store RepositorioMetas) *ServicioMetas {
	return &ServicioMetas{store: store}
}

func (s *ServicioMetas) CrearMeta(ctx context.Context, payload PayloadCrearMeta, empresaID uuid.UUID) (model.Goal, error) {

	goal := &model.Goal{
		Nombre:    payload.Nombre,
		Periodo:   payload.Periodo,
		MontoMeta: payload.MontoMeta,
		EmpresaID: empresaID,
	}

	if err := s.store.CrearMeta(ctx, goal); err != nil {
		return model.Goal{}, err
	}

	return *goal, nil
}

func (s *ServicioMetas) EliminarMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.EliminarMeta(ctx, metaID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *ServicioMetas) ObtenerEstadoMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) (EstadoMeta, error) {

	goal, err := s.store.ObtenerMeta(ctx, metaID, empresaID)

	if err != nil {
		return EstadoMeta{}, err
	}

	if goal.MontoMeta <= 0 {
		return EstadoMeta{}, fmt.Errorf("la meta %s tiene un monto_meta inválido: %d", goal.ID, goal.MontoMeta)
	}

	monto, err := s.store.ObtenerTotalIngresosPorPeriodo(ctx, empresaID, goal.Periodo)

	if err != nil {
		return EstadoMeta{}, err
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

	state := EstadoMeta{
		MontoActual:        monto,
		PorcentajeActual:   porActual,
		MontoPendiente:     pendiente,
		PorcentajeFaltante: porFaltante,
	}

	return state, nil
}

func (s *ServicioMetas) ObtenerMeta(ctx context.Context, metaID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {

	goal, err := s.store.ObtenerMeta(ctx, metaID, empresaID)

	if err != nil {
		return model.Goal{}, err
	}

	return goal, nil
}

func (s *ServicioMetas) ObtenerMetas(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {

	goals, err := s.store.ObtenerMetas(ctx, empresaID)

	if err != nil {
		return nil, err
	}

	return goals, nil
}
