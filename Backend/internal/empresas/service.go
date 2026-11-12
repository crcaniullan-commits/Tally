package empresas

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type RepositorioEmpresas interface {
	Crear(context.Context, *model.Empresa) error
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarEmpresa(context.Context, uuid.UUID, uuid.UUID) error
	ObtenerEmpresas(context.Context, uuid.UUID) ([]model.Empresa, error)
}

type ServicioEmpresas struct {
	store RepositorioEmpresas
}

func NuevoServicioEmpresas(store RepositorioEmpresas) *ServicioEmpresas {
	return &ServicioEmpresas{store}
}

func (s *ServicioEmpresas) ObtenerEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {

	empresas, err := s.store.ObtenerEmpresas(ctx, userID)

	if err != nil {
		return nil, err
	}

	return empresas, nil

}

func (s *ServicioEmpresas) Crear(ctx context.Context, payload PayloadEmpresa) (model.Empresa, error) {
	empresa := &model.Empresa{
		NombreEmpresa: payload.NombreEmpresa,
		UserID:        payload.userID,
	}

	if err := s.store.Crear(ctx, empresa); err != nil {
		return model.Empresa{}, err
	}

	return *empresa, nil

}

func (s *ServicioEmpresas) CambiarNombre(ctx context.Context, nuevoNombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	if err := s.store.CambiarNombre(ctx, nuevoNombre, empresaID, userID); err != nil {
		return err
	}

	return nil
}

func (s *ServicioEmpresas) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	if err := s.store.BorrarEmpresa(ctx, empresaID, userID); err != nil {
		return err
	}

	return nil
}
