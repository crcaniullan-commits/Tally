package debtors

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type RepositorioDeudores interface {
	ObtenerDeudor(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	Crear(context.Context, *model.Debtor) error
	Actualizar(context.Context, uuid.UUID, uuid.UUID, CambiosDeudor) error
	MarcarDeudorPagado(context.Context, uuid.UUID, uuid.UUID) error
	BorrarDeudor(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerDeudores(context.Context, uuid.UUID) ([]model.Debtor, error)
}

// CambiosDeudor son los campos a modificar de una deuda. Descripcion es un
// puntero: nil significa "no tocar la descripción" (el payload de actualización
// no la manda), y un string vacío la deja vacía.
type CambiosDeudor struct {
	Nombre      string
	Monto       int64
	FechaLimite time.Time
	Descripcion *string
}

type ServicioDeudores struct {
	store RepositorioDeudores
}

func NuevoServicioDeudores(store RepositorioDeudores) *ServicioDeudores {
	return &ServicioDeudores{store: store}
}

func (s *ServicioDeudores) Crear(ctx context.Context, empresaID uuid.UUID, payload PayloadDeudor) (model.Debtor, error) {
	debtor := &model.Debtor{
		Nombre:      payload.Nombre,
		Monto:       payload.Monto,
		Descripcion: &payload.Descripcion,
		FechaLimite: &payload.FechaLimite,
		EmpresaID:   empresaID,
	}

	if err := s.store.Crear(ctx, debtor); err != nil {
		return model.Debtor{}, err
	}

	return *debtor, nil
}

func (s *ServicioDeudores) Actualizar(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID, payload PayloadActualizacionDeudor) error {

	cambios := CambiosDeudor{
		Monto:       payload.Monto,
		Nombre:      payload.Nombre,
		FechaLimite: payload.FechaLimite,
		Descripcion: payload.Descripcion,
	}

	if err := s.store.Actualizar(ctx, deudorID, empresaID, cambios); err != nil {
		return err
	}

	return nil
}

// Borrar elimina la deuda de la empresa. Se puede borrar tanto una deuda
// pendiente como una ya pagada.
func (s *ServicioDeudores) Borrar(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.store.BorrarDeudor(ctx, deudorID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *ServicioDeudores) MarcarDeudorPagado(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.MarcarDeudorPagado(ctx, deudorID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *ServicioDeudores) ObtenerDeudores(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	debtors, err := s.store.ObtenerDeudores(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	return debtors, nil
}
