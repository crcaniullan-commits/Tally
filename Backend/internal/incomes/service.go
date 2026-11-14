package incomes

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioIngresos interface {
	AgregarIngreso(context.Context, *model.Income) error
	ActualizarIngreso(context.Context, CambiosIngreso, uuid.UUID, uuid.UUID) (model.Income, error)
	EliminarIngreso(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerIngresosDeEmpresa(context.Context, uuid.UUID, paginacion.ConsultaPaginacionIngresos) ([]model.Income, error)
	CategoriaExiste(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}

// CambiosIngreso son los campos a modificar de un ingreso. Un puntero nil
// significa "no tocar este campo": así el UPDATE puede dejar intacto lo que el
// usuario no mandó.
type CambiosIngreso struct {
	Monto         *int64
	PaymentMethod *model.PaymentMethod
	Descripcion   *string
	CategoryID    *uuid.UUID
	Fecha         *time.Time
}

// Vacio indica que no hay nada que actualizar.
func (c CambiosIngreso) Vacio() bool {
	return c.Monto == nil &&
		c.PaymentMethod == nil &&
		c.Descripcion == nil &&
		c.CategoryID == nil &&
		c.Fecha == nil
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

// ActualizarIngreso modifica un ingreso de la empresa. Solo se tocan los campos
// que vienen en el payload; la fecha se puede corregir (AAAA-MM-DD).
func (s *ServicioIngresos) ActualizarIngreso(ctx context.Context, payload PayloadActualizarIngreso, ingresoID uuid.UUID, empresaID uuid.UUID) (model.Income, error) {
	cambios := CambiosIngreso{
		Monto:         payload.Monto,
		PaymentMethod: payload.PaymentMethod,
		Descripcion:   payload.Descripcion,
		CategoryID:    payload.CategoryID,
	}

	// La fecha se parsea antes de mirar si el payload traía algo: si la fecha
	// vino pero está mal formada, el error es de fecha y no de payload vacío.
	if payload.Fecha != nil {
		fecha, err := time.Parse(time.DateOnly, *payload.Fecha)

		if err != nil {
			return model.Income{}, ErrorFechaInvalida
		}

		cambios.Fecha = &fecha
	}

	if cambios.Vacio() {
		return model.Income{}, ErrorSinCambios
	}

	// La categoría, si viene, tiene que existir y ser global o de la empresa.
	if payload.CategoryID != nil {
		existe, err := s.store.CategoriaExiste(ctx, *payload.CategoryID, empresaID)

		if err != nil {
			return model.Income{}, err
		}

		if !existe {
			return model.Income{}, utilidades.ErrorNoEncontrado
		}
	}

	income, err := s.store.ActualizarIngreso(ctx, cambios, ingresoID, empresaID)

	if err != nil {
		if err == utilidades.ErrorNoEncontrado {
			return model.Income{}, ErrorMovimientoNoEncontrado
		}

		return model.Income{}, err
	}

	return income, nil
}

func (s *ServicioIngresos) ObtenerIngresosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Income, error) {
	incomes, err := s.store.ObtenerIngresosDeEmpresa(ctx, empresaID, fq)
	if err != nil {
		return nil, err
	}

	return incomes, nil
}
