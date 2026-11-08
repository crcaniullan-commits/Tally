package debtors

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// DebtorStoreMock es un mock de DebtorStorage para probar DebtorService sin
// tocar la base de datos.
type DebtorStoreMock struct {
	mock.Mock
}

func (m *DebtorStoreMock) GetDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	args := m.Called(ctx, debtorID, empresaID)
	pagado, _ := args.Get(0).(bool)
	return pagado, args.Error(1)
}

func (m *DebtorStoreMock) Create(ctx context.Context, debtor *model.Debtor) error {
	args := m.Called(ctx, debtor)
	return args.Error(0)
}

func (m *DebtorStoreMock) Update(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID, debtor model.Debtor) error {
	args := m.Called(ctx, debtorID, empresaID, debtor)
	return args.Error(0)
}

func (m *DebtorStoreMock) PaidDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, debtorID, empresaID)
	return args.Error(0)
}

func (m *DebtorStoreMock) GetDebtors(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	args := m.Called(ctx, empresaID)
	debtors, _ := args.Get(0).([]model.Debtor)
	return debtors, args.Error(1)
}

// ServiceDebtorsMock es un mock de ServiceDebtors para probar DebtorsHandler
// de forma aislada.
type ServiceDebtorsMock struct {
	mock.Mock
}

func (m *ServiceDebtorsMock) Create(ctx context.Context, empresaID uuid.UUID, payload DebtorPayload) (model.Debtor, error) {
	args := m.Called(ctx, empresaID, payload)
	debtor, _ := args.Get(0).(model.Debtor)
	return debtor, args.Error(1)
}

func (m *ServiceDebtorsMock) Update(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID, payload UpdatePayload) error {
	args := m.Called(ctx, debtorID, empresaID, payload)
	return args.Error(0)
}

func (m *ServiceDebtorsMock) PaidDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, debtorID, empresaID)
	return args.Error(0)
}

func (m *ServiceDebtorsMock) GetDebtors(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	args := m.Called(ctx, empresaID)
	debtors, _ := args.Get(0).([]model.Debtor)
	return debtors, args.Error(1)
}

var (
	testUserID        = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID     = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testDebtorID      = uuid.MustParse("f6a7b8c9-0d1e-4f2a-8b3c-4d5e6f7a8b9c")
	testOtherDebtorID = uuid.MustParse("a7b8c9d0-1e2f-4a3b-9c4d-5e6f7a8b9c0d")
)

var (
	testDebtorCreatedAt = time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	testFechaLimite     = time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
)

// newTestDebtor construye un model.Debtor válido y reutilizable en los tests.
func newTestDebtor() model.Debtor {
	descripcion := "Deuda pendiente"
	fechaLimite := testFechaLimite

	return model.Debtor{
		ID:          testDebtorID,
		EmpresaID:   testEmpresaID,
		Nombre:      "Cliente",
		Monto:       15000,
		Descripcion: &descripcion,
		Pagado:      false,
		FechaLimite: &fechaLimite,
		CreatedAt:   testDebtorCreatedAt,
	}
}

// newTestDebtorPayload arma el payload de alta con valores que pasan la
// validación.
func newTestDebtorPayload() DebtorPayload {
	return DebtorPayload{
		Nombre:      "Cliente",
		Monto:       15000,
		Descripcion: "Deuda pendiente",
		FechaLimite: testFechaLimite,
	}
}

// newTestUpdatePayload arma el payload de actualización con valores válidos.
func newTestUpdatePayload() UpdatePayload {
	return UpdatePayload{
		Nombre:      "Cliente",
		Monto:       15000,
		FechaLimite: testFechaLimite,
	}
}
