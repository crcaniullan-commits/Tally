package categories

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// CategoriaStoreMock es un mock de la capa de persistencia (RepositorioCategorias) usado
// para probar ServicioCategorias sin tocar la base de datos.
type CategoriaStoreMock struct {
	mock.Mock
}

func (m *CategoriaStoreMock) Crear(ctx context.Context, categoria *model.Categorie) error {
	args := m.Called(ctx, categoria)
	return args.Error(0)
}

func (m *CategoriaStoreMock) CambiarNombre(ctx context.Context, nombre string, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, nombre, categoriaID, empresaID)
	return args.Error(0)
}

func (m *CategoriaStoreMock) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, categoriaID, empresaID)
	return args.Error(0)
}

func (m *CategoriaStoreMock) ObtenerCategoriasPorEmpresa(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	args := m.Called(ctx, empresaID)
	categorias, _ := args.Get(0).([]model.Categorie)
	return categorias, args.Error(1)
}

func (m *CategoriaStoreMock) ObtenerCategoriasVisibles(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	args := m.Called(ctx, empresaID)
	categorias, _ := args.Get(0).([]model.Categorie)
	return categorias, args.Error(1)
}

// IServiceCategoriesMock es un mock de la capa de servicio (IServiceCategorias) usado
// para probar HandlerCategorias de forma aislada.
type IServiceCategoriesMock struct {
	mock.Mock
}

func (m *IServiceCategoriesMock) Crear(ctx context.Context, payload PayloadCategoria, empresaID *uuid.UUID) (model.Categorie, error) {
	args := m.Called(ctx, payload, empresaID)
	categoria, _ := args.Get(0).(model.Categorie)
	return categoria, args.Error(1)
}

func (m *IServiceCategoriesMock) CambiarNombre(ctx context.Context, payload PayloadCategoria, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, payload, categoriaID, empresaID)
	return args.Error(0)
}

func (m *IServiceCategoriesMock) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, categoriaID, empresaID)
	return args.Error(0)
}

func (m *IServiceCategoriesMock) ObtenerCategorias(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	args := m.Called(ctx, empresaID)
	categorias, _ := args.Get(0).([]model.Categorie)
	return categorias, args.Error(1)
}

var (
	testUserID      = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID   = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testCategoriaID = uuid.MustParse("b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d")
)

// newTestCategoria construye un model.Categorie válido y reutilizable en los tests.
func newTestCategoria() model.Categorie {
	empresaID := testEmpresaID
	return model.Categorie{
		ID:        testCategoriaID,
		Nombre:    "Comida",
		EmpresaID: &empresaID,
	}
}

// newTestPayloadCategoria arma el payload de alta con valores que pasan la validación.
func newTestPayloadCategoria() PayloadCategoria {
	return PayloadCategoria{
		Nombre: "Comida",
	}
}
