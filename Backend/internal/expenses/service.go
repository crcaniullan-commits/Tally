package expenses

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioGastos interface {
	AgregarGasto(context.Context, *model.Expense) error
	ActualizarGasto(context.Context, CambiosGasto, uuid.UUID, uuid.UUID) (model.Expense, error)
	EliminarGasto(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerGastosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error)
	CategoriaExiste(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}

// CambiosGasto son los campos a modificar de un gasto. Un puntero nil significa
// "no tocar este campo".
type CambiosGasto struct {
	Monto       *int64
	CategoryID  *uuid.UUID
	Descripcion *string
	Fecha       *time.Time
}

// Vacio indica que no hay nada que actualizar.
func (c CambiosGasto) Vacio() bool {
	return c.Monto == nil &&
		c.CategoryID == nil &&
		c.Descripcion == nil &&
		c.Fecha == nil
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

// ActualizarGasto modifica un gasto de la empresa. Solo se tocan los campos que
// vienen en el payload; la fecha se puede corregir (AAAA-MM-DD).
func (s *ServicioGastos) ActualizarGasto(ctx context.Context, payload PayloadActualizarGasto, gastoID uuid.UUID, empresaID uuid.UUID) (model.Expense, error) {
	cambios := CambiosGasto{
		Monto:       payload.Monto,
		CategoryID:  payload.CategoryID,
		Descripcion: payload.Descripcion,
	}

	// La fecha se parsea antes de mirar si el payload traía algo: si la fecha
	// vino pero está mal formada, el error es de fecha y no de payload vacío.
	if payload.Fecha != nil {
		fecha, err := time.Parse(time.DateOnly, *payload.Fecha)

		if err != nil {
			return model.Expense{}, ErrorFechaInvalida
		}

		cambios.Fecha = &fecha
	}

	if cambios.Vacio() {
		return model.Expense{}, ErrorSinCambios
	}

	// La categoría, si viene, tiene que existir y ser global o de la empresa.
	if payload.CategoryID != nil {
		existe, err := s.store.CategoriaExiste(ctx, *payload.CategoryID, empresaID)

		if err != nil {
			return model.Expense{}, err
		}

		if !existe {
			return model.Expense{}, utilidades.ErrorNoEncontrado
		}
	}

	expense, err := s.store.ActualizarGasto(ctx, cambios, gastoID, empresaID)

	if err != nil {
		if err == utilidades.ErrorNoEncontrado {
			return model.Expense{}, ErrorMovimientoNoEncontrado
		}

		return model.Expense{}, err
	}

	return expense, nil
}

func (s *ServicioGastos) ObtenerGastosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error) {
	expenses, err := s.store.ObtenerGastosDeEmpresa(ctx, empresaID, fq)

	if err != nil {
		return nil, err
	}

	return expenses, nil
}
