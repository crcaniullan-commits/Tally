package empresas

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// StoreEmpresasMock es un mock de la capa de persistencia (StoreEmpresas) usado
// para probar EmpresaService sin tocar la base de datos.
type StoreEmpresasMock struct {
	mock.Mock
}

func (m *StoreEmpresasMock) Create(ctx context.Context, empresa *model.Empresa) error {
	args := m.Called(ctx, empresa)
	return args.Error(0)
}

func (m *StoreEmpresasMock) CambiarNombre(ctx context.Context, nombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, nombre, empresaID, userID)
	return args.Error(0)
}

func (m *StoreEmpresasMock) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, empresaID, userID)
	return args.Error(0)
}

func (m *StoreEmpresasMock) GetEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {
	args := m.Called(ctx, userID)

	empresas, _ := args.Get(0).([]model.Empresa)

	return empresas, args.Error(1)
}

// ServiceEmpresasMock es un mock de la capa de servicio (ServiceEmpresas) usado
// para probar EmpresasHandler de forma aislada.
type ServiceEmpresasMock struct {
	mock.Mock
}

func (m *ServiceEmpresasMock) Create(ctx context.Context, payload empresaPayload) (model.Empresa, error) {
	args := m.Called(ctx, payload)

	empresa, _ := args.Get(0).(model.Empresa)

	return empresa, args.Error(1)
}

func (m *ServiceEmpresasMock) CambiarNombre(ctx context.Context, nombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, nombre, empresaID, userID)
	return args.Error(0)
}

func (m *ServiceEmpresasMock) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, empresaID, userID)
	return args.Error(0)
}

func (m *ServiceEmpresasMock) GetEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {
	args := m.Called(ctx, userID)

	empresas, _ := args.Get(0).([]model.Empresa)

	return empresas, args.Error(1)
}

var (
	testUserID    = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
)

// newTestEmpresa construye un model.Empresa válido y reutilizable en los tests.
func newTestEmpresa() model.Empresa {
	return model.Empresa{
		ID:            testEmpresaID,
		NombreEmpresa: "Mi Empresa",
		UserID:        testUserID,
	}
}

// newTestPayloadEmpresa arma el payload de alta con valores que pasan la validación.
func newTestPayloadEmpresa() empresaPayload {
	return empresaPayload{
		Nombre_empresa: "Mi Empresa",
		user_id:        testUserID,
	}
}
