package incomes

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioIngresos interface {
	AgregarIngreso(context.Context, *model.Income) error
	EliminarIngreso(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerIngresosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Income, error)
	CategoriaExiste(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}
type ServicioIngresos struct {
	store RepositorioIngresos
}

func NuevoServicioIngresos(store RepositorioIngresos) *ServicioIngresos {
	return &ServicioIngresos{store}
}

func (s *ServicioIngresos) AgregarIngreso(ctx context.Context, payload PayloadIngreso, empresaID uuid.UUID) (model.Income, error) {
	income := &model.Income{
		EmpresaID:     empresaID,
		Monto:         payload.Monto,
		PaymentMethod: payload.PaymentMethod,
		Descripcion:   payload.Descripcion,
	}

	// La categoría es opcional (la columna es nullable). Si viene, tiene que
	// existir y ser global o de la empresa; una categoría de otra empresa se
	// rechaza igual que una inexistente.
	if payload.CategoryID != nil {
		existe, err := s.store.CategoriaExiste(ctx, *payload.CategoryID, empresaID)
		if err != nil {
			return model.Income{}, err
		}
		if !existe {
			return model.Income{}, utilidades.ErrorNoEncontrado
		}

		income.CategoryID = payload.CategoryID
	}

	if err := s.store.AgregarIngreso(ctx, income); err != nil {
		return model.Income{}, err
	}

	return *income, nil
}

func (s *ServicioIngresos) EliminarIngreso(ctx context.Context, ingresoID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.store.EliminarIngreso(ctx, ingresoID, empresaID); err != nil {
		return err
	}
	return nil
}

func (s *ServicioIngresos) ObtenerIngresosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Income, error) {
	incomes, err := s.store.ObtenerIngresosDeEmpresa(ctx, empresaID, fq)
	if err != nil {
		return nil, err
	}

	return incomes, nil
}
