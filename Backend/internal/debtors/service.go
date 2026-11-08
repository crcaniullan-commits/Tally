package debtors

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type DebtorStorage interface {
	GetDebtor(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	Create(context.Context, *model.Debtor) error
	Update(context.Context, uuid.UUID, uuid.UUID, model.Debtor) error
	PaidDebtor(context.Context, uuid.UUID, uuid.UUID) error
	GetDebtors(context.Context, uuid.UUID) ([]model.Debtor, error)
}
type DebtorService struct {
	store DebtorStorage
}

func NewDebtorService(store DebtorStorage) *DebtorService {
	return &DebtorService{store: store}
}

func (s *DebtorService) Create(ctx context.Context, empresaID uuid.UUID, payload DebtorPayload) (model.Debtor, error) {
	debtor := &model.Debtor{
		Nombre:      payload.Nombre,
		Monto:       payload.Monto,
		Descripcion: &payload.Descripcion,
		FechaLimite: &payload.FechaLimite,
		EmpresaID:   empresaID,
	}

	if err := s.store.Create(ctx, debtor); err != nil {
		return model.Debtor{}, err
	}

	return *debtor, nil
}

func (s *DebtorService) Update(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID, payload UpdatePayload) error {

	debtor := model.Debtor{
		Monto:       payload.Monto,
		Nombre:      payload.Nombre,
		FechaLimite: &payload.FechaLimite,
	}

	if err := s.store.Update(ctx, debtorID, empresaID, debtor); err != nil {
		return err
	}

	return nil
}

func (s *DebtorService) PaidDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.PaidDebtor(ctx, debtorID, empresaID); err != nil {
		return err
	}

	return nil
}

func (s *DebtorService) GetDebtors(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	debtors, err := s.store.GetDebtors(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	return debtors, nil
}
