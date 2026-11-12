package expenses

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioGastos interface {
	AgregarGasto(context.Context, *model.Expense) error
	EliminarGasto(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerGastosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error)
	CategoriaExiste(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}

type ServicioGastos struct {
	store RepositorioGastos
}

func NuevoServicioGastos(store RepositorioGastos) *ServicioGastos {
	return &ServicioGastos{store: store}
}

func (s *ServicioGastos) AgregarGasto(ctx context.Context, payload PayloadGasto, empresaID uuid.UUID) (model.Expense, error) {
	expense := &model.Expense{
		EmpresaID:   empresaID,
		Monto:       payload.Monto,
		Descripcion: payload.Descripcion,
	}

	if payload.CategoryID == nil {
		return model.Expense{}, utilidades.ErrorNoEncontrado
	}

	existe, err := s.store.CategoriaExiste(ctx, *payload.CategoryID, empresaID)
	if err != nil {
		return model.Expense{}, err
	}
	if !existe {
		return model.Expense{}, utilidades.ErrorNoEncontrado
	}

	expense.CategoryID = *payload.CategoryID

	if err := s.store.AgregarGasto(ctx, expense); err != nil {
		return model.Expense{}, err
	}

	return *expense, nil
}

func (s *ServicioGastos) EliminarGasto(ctx context.Context, gastoID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.store.EliminarGasto(ctx, gastoID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *ServicioGastos) ObtenerGastosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error) {
	expenses, err := s.store.ObtenerGastosDeEmpresa(ctx, empresaID, fq)

	if err != nil {
		return nil, err
	}

	return expenses, nil
}
