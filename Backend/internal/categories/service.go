package categories

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type CategoriaStore interface {
	Create(context.Context, *model.Categorie) error
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarCategoria(context.Context, uuid.UUID, uuid.UUID) error
	GetCategoriaPorNombreYEmpresa(context.Context, model.Categorie) (bool, error)
}
type CategorieService struct {
	store CategoriaStore
}

func NewCategorieService(store CategoriaStore) *CategorieService {
	return &CategorieService{store: store}
}

func (s *CategorieService) Create(ctx context.Context, payload CategoriaPayload, categoriaID *uuid.UUID) (model.Categorie, error) {
	categoria := &model.Categorie{
		Nombre:    payload.Nombre,
		EmpresaID: categoriaID,
	}

	exist, err := s.store.GetCategoriaPorNombreYEmpresa(ctx, *categoria)

	if exist {
		if err != nil {
			return model.Categorie{}, err
		}
	}

	if err = s.store.Create(ctx, categoria); err != nil {
		return model.Categorie{}, err
	}

	return *categoria, nil
}

func (s *CategorieService) CambiarNombre(ctx context.Context, payload CategoriaPayload, categoriaID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.CambiarNombre(ctx, payload.Nombre, categoriaID, empresaID); err != nil {
		return err
	}
	return nil
}

func (s *CategorieService) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {

	if err := s.store.BorrarCategoria(ctx, categoriaID, empresaID); err != nil {
		return err
	}
	return nil
}
