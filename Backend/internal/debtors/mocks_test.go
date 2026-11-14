package debtors

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// AlmacenDeudoresMock es un mock de RepositorioDeudores para probar ServicioDeudores sin
// tocar la base de datos.
type AlmacenDeudoresMock struct {
	mock.Mock
}

func (m *AlmacenDeudoresMock) ObtenerDeudor(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	args := m.Called(ctx, deudorID, empresaID)
	pagado, _ := args.Get(0).(bool)
	return pagado, args.Error(1)
}

func (m *AlmacenDeudoresMock) Crear(ctx context.Context, debtor *model.Debtor) error {
	args := m.Called(ctx, debtor)
	return args.Error(0)
}

func (m *AlmacenDeudoresMock) Actualizar(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID, debtor model.Debtor) error {
	args := m.Called(ctx, deudorID, empresaID, debtor)
	return args.Error(0)
}

func (m *AlmacenDeudoresMock) MarcarDeudorPagado(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, deudorID, empresaID)
	return args.Error(0)
}

func (m *AlmacenDeudoresMock) ObtenerDeudores(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	args := m.Called(ctx, empresaID)
	debtors, _ := args.Get(0).([]model.Debtor)
	return debtors, args.Error(1)
}

// IServiceDeudoresMock es un mock de IServiceDeudores para probar HandlerDeudores
// de forma aislada.
type IServiceDeudoresMock struct {
	mock.Mock
}

func (m *IServiceDeudoresMock) Crear(ctx context.Context, empresaID uuid.UUID, payload PayloadDeudor) (model.Debtor, error) {
	args := m.Called(ctx, empresaID, payload)
	debtor, _ := args.Get(0).(model.Debtor)
	return debtor, args.Error(1)
}

func (m *IServiceDeudoresMock) Actualizar(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID, payload PayloadActualizacionDeudor) error {
	args := m.Called(ctx, deudorID, empresaID, payload)
	return args.Error(0)
}

func (m *IServiceDeudoresMock) MarcarDeudorPagado(ctx context.Context, deudorID uuid.UUID, empresaID uuid.UUID) error {
	args := m.Called(ctx, deudorID, empresaID)
	return args.Error(0)
}

func (m *IServiceDeudoresMock) ObtenerDeudores(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
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
func newTestDebtorPayload() PayloadDeudor {
	return PayloadDeudor{
		Nombre:      "Cliente",
		Monto:       15000,
		Descripcion: "Deuda pendiente",
		FechaLimite: testFechaLimite,
	}
}

// newTestUpdatePayload arma el payload de actualización con valores válidos.
func newTestUpdatePayload() PayloadActualizacionDeudor {
	return PayloadActualizacionDeudor{
		Nombre:      "Cliente",
		Monto:       15000,
		FechaLimite: testFechaLimite,
	}
}
