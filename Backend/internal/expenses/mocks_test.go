package expenses

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// RepositorioExpensesMock es un mock de la capa de persistencia (RepositorioGastos) usado
// para probar ServicioGastos sin tocar la base de datos.
type RepositorioExpensesMock struct {
	mock.Mock
}

func (m *RepositorioExpensesMock) AgregarGasto(ctx context.Context, expense *model.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *RepositorioExpensesMock) EliminarGasto(ctx context.Context, gastoID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, gastoID, empresaID)
	return args.Error(0)
}

func (m *RepositorioExpensesMock) ObtenerGastosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error) {
	args := m.Called(ctx, empresaID, fq)

	expenses, _ := args.Get(0).([]model.Expense)

	return expenses, args.Error(1)
}

func (m *RepositorioExpensesMock) CategoriaExiste(ctx context.Context, categoryID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID, empresaID)

	existe, _ := args.Get(0).(bool)

	return existe, args.Error(1)
}

// IServiceExpensesMock es un mock de la capa de servicio (IServiceGastos) usado
// para probar HandlerGastos de forma aislada.
type IServiceExpensesMock struct {
	mock.Mock
}

func (m *IServiceExpensesMock) AgregarGasto(ctx context.Context, payload PayloadGasto, empresaID uuid.UUID) (model.Expense, error) {
	args := m.Called(ctx, payload, empresaID)

	expense, _ := args.Get(0).(model.Expense)

	return expense, args.Error(1)
}

func (m *IServiceExpensesMock) EliminarGasto(ctx context.Context, gastoID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, gastoID, empresaID)
	return args.Error(0)
}

func (m *IServiceExpensesMock) ObtenerGastosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error) {
	args := m.Called(ctx, empresaID, fq)

	expenses, _ := args.Get(0).([]model.Expense)

	return expenses, args.Error(1)
}

var (
	testUserID       = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID    = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testExpenseID    = uuid.MustParse("9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e")
	testCategoryID   = uuid.MustParse("b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d")
	testCategoryName = "Comida"
	testFecha        = time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
)

// newTestExpense construye un model.Expense válido y reutilizable en los tests.
func newTestExpense() model.Expense {
	descripcion := "Compra insumos"
	return model.Expense{
		ID:          testExpenseID,
		EmpresaID:   testEmpresaID,
		Monto:       1500,
		CategoryID:  testCategoryID,
		Descripcion: &descripcion,
		Fecha:       testFecha,
		CreatedAt:   testFecha.Add(2 * time.Hour),
	}
}

// newTestPayloadExpense arma el payload de alta con valores que pasan la validación.
func newTestPayloadExpense() PayloadGasto {
	descripcion := "Compra insumos"
	return PayloadGasto{
		Monto:       1500,
		CategoryID:  &testCategoryID,
		Descripcion: &descripcion,
	}
}

// newTestFilterQuery es la paginación por defecto.
func newTestFilterQuery() paginacion.ConsultaPaginacionIngresos {
	return paginacion.NuevaConsultaPaginacionIngresos()
}
