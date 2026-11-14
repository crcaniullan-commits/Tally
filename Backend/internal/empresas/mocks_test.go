package empresas

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// RepositorioEmpresasMock es un mock de la capa de persistencia (RepositorioEmpresas) usado
// para probar ServicioEmpresas sin tocar la base de datos.
type RepositorioEmpresasMock struct {
	mock.Mock
}

func (m *RepositorioEmpresasMock) Crear(ctx context.Context, empresa *model.Empresa) error {
	args := m.Called(ctx, empresa)
	return args.Error(0)
}

func (m *RepositorioEmpresasMock) CambiarNombre(ctx context.Context, nombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, nombre, empresaID, userID)
	return args.Error(0)
}

func (m *RepositorioEmpresasMock) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, empresaID, userID)
	return args.Error(0)
}

func (m *RepositorioEmpresasMock) ObtenerEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {
	args := m.Called(ctx, userID)

	empresas, _ := args.Get(0).([]model.Empresa)

	return empresas, args.Error(1)
}

// IServiceEmpresasMock es un mock de la capa de servicio (IServiceEmpresas) usado
// para probar HandlerEmpresas de forma aislada.
type IServiceEmpresasMock struct {
	mock.Mock
}

func (m *IServiceEmpresasMock) Crear(ctx context.Context, payload PayloadEmpresa) (model.Empresa, error) {
	args := m.Called(ctx, payload)

	empresa, _ := args.Get(0).(model.Empresa)

	return empresa, args.Error(1)
}

func (m *IServiceEmpresasMock) CambiarNombre(ctx context.Context, nombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, nombre, empresaID, userID)
	return args.Error(0)
}

func (m *IServiceEmpresasMock) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, empresaID, userID)
	return args.Error(0)
}

func (m *IServiceEmpresasMock) ObtenerEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {
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
func newTestPayloadEmpresa() PayloadEmpresa {
	return PayloadEmpresa{
		NombreEmpresa: "Mi Empresa",
		userID:        testUserID,
	}
}
