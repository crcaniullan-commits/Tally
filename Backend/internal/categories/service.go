package categories

import (
	"context"
	"strings"
	"unicode"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

type CategoriaStore interface {
	Create(context.Context, *model.Categorie) error
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarCategoria(context.Context, uuid.UUID, uuid.UUID) error
	GetCategoriasPorEmpresa(context.Context, uuid.UUID) ([]model.Categorie, error)
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

	if categoria.EmpresaID != nil {
		if err := s.validarNombreUnico(ctx, categoria.Nombre, *categoria.EmpresaID, nil); err != nil {
			return model.Categorie{}, err
		}
	}

	if err := s.store.Create(ctx, categoria); err != nil {
		return model.Categorie{}, err
	}

	return *categoria, nil
}

func (s *CategorieService) CambiarNombre(ctx context.Context, payload CategoriaPayload, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	if err := s.validarNombreUnico(ctx, payload.Nombre, empresaID, &categoriaID); err != nil {
		return err
	}

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

// normalizarNombre deja el nombre comparable: sin espacios en los extremos, en
// minúsculas y sin acentos. "Comida", "comida" y "COMIDA" normalizan igual, y
// también "Áqui", "aqui" y "áqui".
func normalizarNombre(nombre string) string {
	descompuesto := norm.NFD.String(strings.TrimSpace(nombre))

	var normalizado strings.Builder
	normalizado.Grow(len(descompuesto))

	for _, r := range descompuesto {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		normalizado.WriteRune(r)
	}

	return strings.ToLower(normalizado.String())
}

// validarNombreUnico devuelve AlreadyExistError si otra categoría de la misma
// empresa ya usa el nombre, sin distinguir mayúsculas, acentos ni espacios.
// excluirID permite ignorar la categoría que se está renombrando.
func (s *CategorieService) validarNombreUnico(ctx context.Context, nombre string, empresaID uuid.UUID, excluirID *uuid.UUID) error {
	categorias, err := s.store.GetCategoriasPorEmpresa(ctx, empresaID)
	if err != nil {
		return err
	}

	buscado := normalizarNombre(nombre)

	for _, categoria := range categorias {
		if excluirID != nil && categoria.ID == *excluirID {
			continue
		}
		if normalizarNombre(categoria.Nombre) == buscado {
			return AlreadyExistError
		}
	}

	return nil
}
