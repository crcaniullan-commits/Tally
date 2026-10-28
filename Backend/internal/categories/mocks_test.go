package categories

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// CategoriaStoreMock es un mock de la capa de persistencia (CategoriaStore) usado
// para probar CategorieService sin tocar la base de datos.
type CategoriaStoreMock struct {
	mock.Mock
}

func (m *CategoriaStoreMock) Create(ctx context.Context, categoria *model.Categorie) error {
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

func (m *CategoriaStoreMock) GetCategoriaPorNombreYEmpresa(ctx context.Context, categoria model.Categorie) (bool, error) {
	args := m.Called(ctx, categoria)
	exist, _ := args.Get(0).(bool)
	return exist, args.Error(1)
}

// ServiceCategoriesMock es un mock de la capa de servicio (ServiceCategories) usado
// para probar CategoriesHandler de forma aislada.
type ServiceCategoriesMock struct {
	mock.Mock
}

func (m *ServiceCategoriesMock) Create(ctx context.Context, payload CategoriaPayload, empresaID *uuid.UUID) (model.Categorie, error) {
	args := m.Called(ctx, payload, empresaID)
	categoria, _ := args.Get(0).(model.Categorie)
	return categoria, args.Error(1)
}

func (m *ServiceCategoriesMock) CambiarNombre(ctx context.Context, payload CategoriaPayload, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, payload, categoriaID, empresaID)
	return args.Error(0)
}

func (m *ServiceCategoriesMock) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, categoriaID, empresaID)
	return args.Error(0)
}

var (
	testUserID     = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID  = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
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
func newTestPayloadCategoria() CategoriaPayload {
	return CategoriaPayload{
		Nombre: "Comida",
	}
}
