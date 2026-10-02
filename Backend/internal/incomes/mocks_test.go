package incomes

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// StoreIncomesMock es un mock de la capa de persistencia (StoreIncomes) usado
// para probar IncomeService sin tocar la base de datos.
type StoreIncomesMock struct {
	mock.Mock
}

func (m *StoreIncomesMock) AddIncome(ctx context.Context, income *IncomeStorage) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *StoreIncomesMock) DeleteIncome(ctx context.Context, incomeID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, incomeID, userID)
	return args.Error(0)
}

// GetAllIncomesOfUser pasa la paginación a m.Called como tercer argumento para
// que los tests puedan assertear los query params que llegaron al store.
func (m *StoreIncomesMock) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID, fq pagination.IncomePaginationQuery) ([]IncomeStorage, error) {
	args := m.Called(ctx, userID, fq)

	incomes, _ := args.Get(0).([]IncomeStorage)

	return incomes, args.Error(1)
}

// ServiceIncomesMock es un mock de la capa de servicio (ServiceIncomes) usado
// para probar IncomesHandler de forma aislada.
type ServiceIncomesMock struct {
	mock.Mock
}

func (m *ServiceIncomesMock) AddIncome(ctx context.Context, payload IncomePayload, userID uuid.UUID) (IncomeStorage, error) {
	args := m.Called(ctx, payload, userID)

	income, _ := args.Get(0).(IncomeStorage)

	return income, args.Error(1)
}

func (m *ServiceIncomesMock) DeleteIncome(ctx context.Context, incomeID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, incomeID, userID)
	return args.Error(0)
}

func (m *ServiceIncomesMock) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID, fq pagination.IncomePaginationQuery) ([]IncomeStorage, error) {
	args := m.Called(ctx, userID, fq)

	incomes, _ := args.Get(0).([]IncomeStorage)

	return incomes, args.Error(1)
}

var (
	testUserID     = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testIncomeID   = uuid.MustParse("9c8b7a65-4d3c-4b2a-9f8e-7d6c5b4a3f2e")
	testCategoryID = uuid.MustParse("b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d")
	testFecha      = time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
)

// newTestIncome construye un IncomeStorage válido y reutilizable en los tests.
func newTestIncome() IncomeStorage {
	descripcion := "Venta de almuerzo"

	return IncomeStorage{
		ID:            testIncomeID,
		UserID:        testUserID,
		Monto:         1500,
		PaymentMethod: util.PaymentMethodDebito,
		Descripcion:   &descripcion,
		Fecha:         testFecha,
		CreatedAt:     testFecha.Add(2 * time.Hour),
		CategoryID:    &testCategoryID,
	}
}

// newTestPayload arma el payload de alta con valores que pasan la validación.
func newTestPayload() IncomePayload {
	descripcion := "Venta de almuerzo"

	return IncomePayload{
		Monto:         1500,
		PaymentMethod: util.PaymentMethodDebito,
		Descripcion:   &descripcion,
		CategoryID:    &testCategoryID,
	}
}

// newTestFilterQuery es la paginación por defecto (primera página, sin rango de
// fechas ni categoría), que es lo que el handler manda cuando la query string
// viene vacía. Los tests que agregan ?limit / ?offset / ?since / ?until /
// ?category_id arman la suya a mano.
func newTestFilterQuery() pagination.IncomePaginationQuery {
	return pagination.NewIncomePaginationQuery()
}

// filterQueryWithRango es una paginación con filtros de fecha, para probar que
// llegan intactos hasta el store.
func filterQueryWithRango() pagination.IncomePaginationQuery {
	return pagination.IncomePaginationQuery{
		Limit:  5,
		Offset: 10,
		Since:  "2026-09-01",
		Until:  "2026-09-30",
	}
}

// filterQueryWithCategory es una paginación filtrando por categoría: el
// category_id viaja como string porque así lo recibe el SELECT ($6 = ” OR
// category_id = $6::uuid).
func filterQueryWithCategory() pagination.IncomePaginationQuery {
	return pagination.IncomePaginationQuery{
		Limit:      pagination.DefaultLimit,
		CategoryID: testCategoryID.String(),
	}
}
