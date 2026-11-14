package incomes

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// RepositorioIngresosMock es un mock de la capa de persistencia (RepositorioIngresos) usado
// para probar ServicioIngresos sin tocar la base de datos.
type RepositorioIngresosMock struct {
	mock.Mock
}

func (m *RepositorioIngresosMock) AgregarIngreso(ctx context.Context, income *model.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *RepositorioIngresosMock) EliminarIngreso(ctx context.Context, ingresoID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, ingresoID, empresaID)
	return args.Error(0)
}

func (m *RepositorioIngresosMock) ActualizarIngreso(ctx context.Context, cambios CambiosIngreso, ingresoID uuid.UUID, empresaID uuid.UUID) (model.Income, error) {
	args := m.Called(ctx, cambios, ingresoID, empresaID)

	income, _ := args.Get(0).(model.Income)

	return income, args.Error(1)
}

// ObtenerIngresosDeEmpresa pasa la paginación a m.Called como tercer argumento para
// que los tests puedan assertear los query params que llegaron al store.
func (m *RepositorioIngresosMock) ObtenerIngresosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Income, error) {
	args := m.Called(ctx, empresaID, fq)

	incomes, _ := args.Get(0).([]model.Income)

	return incomes, args.Error(1)
}

// CategoriaExiste indica si la categoría es visible para la empresa. El mock
// devuelve lo que le configure el test.
func (m *RepositorioIngresosMock) CategoriaExiste(ctx context.Context, categoryID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID, empresaID)

	existe, _ := args.Get(0).(bool)

	return existe, args.Error(1)
}

// IServiceIncomesMock es un mock de la capa de servicio (IServiceIngresos) usado
// para probar HandlerIngresos de forma aislada.
type IServiceIncomesMock struct {
	mock.Mock
}

func (m *IServiceIncomesMock) AgregarIngreso(ctx context.Context, payload PayloadIngreso, empresaID uuid.UUID) (model.Income, error) {
	args := m.Called(ctx, payload, empresaID)

	income, _ := args.Get(0).(model.Income)

	return income, args.Error(1)
}

func (m *IServiceIncomesMock) EliminarIngreso(ctx context.Context, ingresoID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, ingresoID, empresaID)
	return args.Error(0)
}

func (m *IServiceIncomesMock) ActualizarIngreso(ctx context.Context, payload PayloadActualizarIngreso, ingresoID uuid.UUID, empresaID uuid.UUID) (model.Income, error) {
	args := m.Called(ctx, payload, ingresoID, empresaID)

	income, _ := args.Get(0).(model.Income)

	return income, args.Error(1)
}

func (m *IServiceIncomesMock) ObtenerIngresosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Income, error) {
	args := m.Called(ctx, empresaID, fq)

	incomes, _ := args.Get(0).([]model.Income)

	return incomes, args.Error(1)
}

var (
	testUserID       = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID    = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testIncomeID     = uuid.MustParse("9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e")
	testCategoryID   = uuid.MustParse("b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d")
	testCategoryName = "Comida"
	testFecha        = time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
)

// newTestIncome construye un model.Income válido y reutilizable en los tests.
func newTestIncome() model.Income {
	descripcion := "Venta de almuerzo"

	return model.Income{
		ID:            testIncomeID,
		EmpresaID:     testEmpresaID,
		Monto:         1500,
		PaymentMethod: utilidades.MetodoPagoDebito,
		Descripcion:   &descripcion,
		Fecha:         testFecha,
		CreatedAt:     testFecha.Add(2 * time.Hour),
		CategoryID:    &testCategoryID,
	}
}

// newTestPayload arma el payload de alta con valores que pasan la validación.
func newTestPayload() PayloadIngreso {
	descripcion := "Venta de almuerzo"

	return PayloadIngreso{
		Monto:         1500,
		PaymentMethod: utilidades.MetodoPagoDebito,
		Descripcion:   &descripcion,
		CategoryID:    &testCategoryID,
	}
}

// newTestFilterQuery es la paginación por defecto (primera página, sin rango de
// fechas ni categoría), que es lo que el handler manda cuando la query string
// viene vacía. Los tests que agregan ?limit / ?offset / ?since / ?until /
// ?category_id arman la suya a mano.
func newTestFilterQuery() paginacion.ConsultaPaginacionIngresos {
	return paginacion.NuevaConsultaPaginacionIngresos()
}

// filterQueryWithRango es una paginación con filtros de fecha, para probar que
// llegan intactos hasta el store.
func filterQueryWithRango() paginacion.ConsultaPaginacionIngresos {
	return paginacion.ConsultaPaginacionIngresos{
		Limite:         5,
		Desplazamiento: 10,
		Desde:          "2026-09-01",
		Hasta:          "2026-09-30",
	}
}

// filterQueryWithCategory es una paginación filtrando por categoría: el
// category_id viaja como string porque así lo recibe el SELECT ($6 = ” OR
// category_id = $6::uuid).
func filterQueryWithCategory() paginacion.ConsultaPaginacionIngresos {
	return paginacion.ConsultaPaginacionIngresos{
		Limite:     paginacion.LimitePorDefecto,
		CategoryID: testCategoryID.String(),
	}
}
