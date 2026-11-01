package empresas

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type StoreEmpresas interface {
	Create(context.Context, *model.Empresa) error
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarEmpresa(context.Context, uuid.UUID, uuid.UUID) error
	GetEmpresas(context.Context, uuid.UUID) ([]model.Empresa, error)
}

type EmpresaService struct {
	store StoreEmpresas
}

func NewEmpresaService(store StoreEmpresas) *EmpresaService {
	return &EmpresaService{store}
}

func (s *EmpresaService) GetEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {

	empresas, err := s.store.GetEmpresas(ctx, userID)

	if err != nil {
		return nil, err
	}

	return empresas, nil

}

func (s *EmpresaService) Create(ctx context.Context, payload empresaPayload) (model.Empresa, error) {
	empresa := &model.Empresa{
		NombreEmpresa: payload.Nombre_empresa,
		UserID:        payload.user_id,
	}

	if err := s.store.Create(ctx, empresa); err != nil {
		return model.Empresa{}, err
	}

	return *empresa, nil

}

func (s *EmpresaService) CambiarNombre(ctx context.Context, nuevoNombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	if err := s.store.CambiarNombre(ctx, nuevoNombre, empresaID, userID); err != nil {
		return err
	}

	return nil
}

func (s *EmpresaService) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	if err := s.store.BorrarEmpresa(ctx, empresaID, userID); err != nil {
		return err
	}

	return nil
}
