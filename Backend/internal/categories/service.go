package categories

import (
	"context"
	"strings"
	"unicode"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

type RepositorioCategorias interface {
	Crear(context.Context, *model.Categorie) error
	CambiarNombre(context.Context, string, uuid.UUID, uuid.UUID) error
	BorrarCategoria(context.Context, uuid.UUID, uuid.UUID) error
	TieneMovimientos(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	ObtenerCategoriasPorEmpresa(context.Context, uuid.UUID) ([]model.Categorie, error)
	ObtenerCategoriasVisibles(context.Context, uuid.UUID) ([]model.Categorie, error)
}
type ServicioCategorias struct {
	store RepositorioCategorias
}

func NuevoServicioCategorias(store RepositorioCategorias) *ServicioCategorias {
	return &ServicioCategorias{store: store}
}

func (s *ServicioCategorias) Crear(ctx context.Context, payload PayloadCategoria, empresaID *uuid.UUID) (model.Categorie, error) {
	if !esTipoValido(payload.Tipo) {
		return model.Categorie{}, ErrorTipoCategoriaInvalido
	}

	categoria := &model.Categorie{
		Nombre:    payload.Nombre,
		Tipo:      payload.Tipo,
		EmpresaID: empresaID,
	}

	if categoria.EmpresaID != nil {
		if err := s.validarNombreUnico(ctx, categoria.Nombre, *categoria.EmpresaID, &categoria.Tipo, nil); err != nil {
			return model.Categorie{}, err
		}
	}

	if err := s.store.Crear(ctx, categoria); err != nil {
		return model.Categorie{}, err
	}

	return *categoria, nil
}

func (s *ServicioCategorias) CambiarNombre(ctx context.Context, payload PayloadCambiarNombre, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	// El tipo no se renombra: la unicidad del nombre se evalúa contra las
	// categorías del mismo tipo que la que se está renombrando.
	if err := s.validarNombreUnico(ctx, payload.Nombre, empresaID, nil, &categoriaID); err != nil {
		return err
	}

	if err := s.store.CambiarNombre(ctx, payload.Nombre, categoriaID, empresaID); err != nil {
		return err
	}
	return nil
}

func (s *ServicioCategorias) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	// Una categoría con movimientos no se puede borrar: expenses.category_id es
	// NOT NULL con ON DELETE CASCADE (el borrado se llevaría los gastos) e
	// incomes.category_id tiene ON DELETE SET NULL (los ingresos quedarían sin
	// categoría). Es mejor un 409 explícito que perder movimientos en silencio.
	tieneMovimientos, err := s.store.TieneMovimientos(ctx, categoriaID, empresaID)

	if err != nil {
		return err
	}

	if tieneMovimientos {
		return ErrorCategoriaConMovimientos
	}

	if err := s.store.BorrarCategoria(ctx, categoriaID, empresaID); err != nil {
		return err
	}
	return nil
}

func (s *ServicioCategorias) ObtenerCategorias(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	return s.store.ObtenerCategoriasVisibles(ctx, empresaID)
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

// validarNombreUnico devuelve ErrorCategoriaDuplicada si otra categoría de la
// misma empresa ya usa el nombre, sin distinguir mayúsculas, acentos ni espacios.
// excluirID permite ignorar la categoría que se está renombrando.
//
// El nombre solo tiene que ser único dentro del mismo tipo: "Comida" puede
// existir como categoría de gastos y como categoría de ingresos de la misma
// empresa. Si tipo es nil se usa el de la categoría que se está renombrando
// (excluirID); si esa categoría no existe, no hay nada que validar.
func (s *ServicioCategorias) validarNombreUnico(ctx context.Context, nombre string, empresaID uuid.UUID, tipo *model.CategoryType, excluirID *uuid.UUID) error {
	categorias, err := s.store.ObtenerCategoriasPorEmpresa(ctx, empresaID)
	if err != nil {
		return err
	}

	if tipo == nil {
		for _, categoria := range categorias {
			if excluirID != nil && categoria.ID == *excluirID {
				tipo = &categoria.Tipo
				break
			}
		}
	}

	// La categoría a renombrar no existe: el store responde ErrorNoEncontrado.
	if tipo == nil {
		return nil
	}

	buscado := normalizarNombre(nombre)

	for _, categoria := range categorias {
		if excluirID != nil && categoria.ID == *excluirID {
			continue
		}
		if categoria.Tipo != *tipo {
			continue
		}
		if normalizarNombre(categoria.Nombre) == buscado {
			return ErrorCategoriaDuplicada
		}
	}

	return nil
}

// esTipoValido indica si el tipo es uno de los dos que admite la base.
func esTipoValido(tipo model.CategoryType) bool {
	return tipo == model.CategoryTypeIngreso || tipo == model.CategoryTypeGasto
}
