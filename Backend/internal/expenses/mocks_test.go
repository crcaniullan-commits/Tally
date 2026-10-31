package expenses

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// StoreExpensesMock es un mock de la capa de persistencia (StoreExpenses) usado
// para probar ExpenseService sin tocar la base de datos.
type StoreExpensesMock struct {
	mock.Mock
}

func (m *StoreExpensesMock) AddExpense(ctx context.Context, expense *model.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *StoreExpensesMock) DeleteExpense(ctx context.Context, expenseID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, expenseID, empresaID)
	return args.Error(0)
}

func (m *StoreExpensesMock) GetAllExpensesOfEmpresa(ctx context.Context, empresaID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Expense, error) {
	args := m.Called(ctx, empresaID, fq)

	expenses, _ := args.Get(0).([]model.Expense)

	return expenses, args.Error(1)
}

func (m *StoreExpensesMock) CategoryExists(ctx context.Context, categoryID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID, empresaID)

	existe, _ := args.Get(0).(bool)

	return existe, args.Error(1)
}

// ServiceExpensesMock es un mock de la capa de servicio (ServiceExpenses) usado
// para probar ExpensesHandler de forma aislada.
type ServiceExpensesMock struct {
	mock.Mock
}

func (m *ServiceExpensesMock) AddExpense(ctx context.Context, payload ExpensePayload, empresaID uuid.UUID) (model.Expense, error) {
	args := m.Called(ctx, payload, empresaID)

	expense, _ := args.Get(0).(model.Expense)

	return expense, args.Error(1)
}

func (m *ServiceExpensesMock) DeleteExpense(ctx context.Context, expenseID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, expenseID, empresaID)
	return args.Error(0)
}

func (m *ServiceExpensesMock) GetAllExpensesOfEmpresa(ctx context.Context, empresaID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Expense, error) {
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
func newTestPayloadExpense() ExpensePayload {
	descripcion := "Compra insumos"
	return ExpensePayload{
		Monto:       1500,
		CategoryID:  &testCategoryID,
		Descripcion: &descripcion,
	}
}

// newTestFilterQuery es la paginación por defecto.
func newTestFilterQuery() pagination.IncomePaginationQuery {
	return pagination.NewIncomePaginationQuery()
}
