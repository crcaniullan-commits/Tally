package incomes

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestIncomeService_AddIncome(t *testing.T) {
	t.Run("mapea el payload a model.Income con el ID de la empresa", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		payload := newTestPayload()

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(testCategoryID, nil).Once()

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.EmpresaID == testEmpresaID &&
					i.Monto == payload.Monto &&
					i.PaymentMethod == util.PaymentMethodDebito &&
					i.Descripcion != nil && *i.Descripcion == "Venta de almuerzo" &&
					i.CategoryID != nil && *i.CategoryID == testCategoryID
			}),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertNumberOfCalls(t, "AddIncome", 1)
	})

	t.Run("resuelve el nombre de la categoria al id antes de persistir", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(testCategoryID, nil).Once()
		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.CategoryID != nil && *i.CategoryID == testCategoryID
			}),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testEmpresaID)

		require.NoError(t, err)
		store.AssertCalled(t, "GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID)
		store.AssertExpectations(t)
	})

	t.Run("acepta descripcion nula", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		payload := IncomePayload{Monto: 300, PaymentMethod: util.PaymentMethodEfectivo}

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool { return i.Descripcion == nil }),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("deja la categoria en null si el payload no trae categoria", func(t *testing.T) {
		// incomes.category_id es nullable (migrations/000006). El service no
		// tiene que inventar una categoria por defecto: un nil tiene que
		// seguir siendo nil para que el INSERT escriba NULL y no un UUID cero
		// que no matchearia ninguna fila de categories.
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool { return i.CategoryID == nil }),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(),
			IncomePayload{Monto: 300, PaymentMethod: util.PaymentMethodEfectivo}, testEmpresaID)

		require.NoError(t, err)
		store.AssertNotCalled(t, "GetCategoryIDByName", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound si la categoria no existe y no persiste el ingreso", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(uuid.Nil, util.ErrNotFound).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store de categorias sin persistir", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(uuid.Nil, errStoreBoom).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertNotCalled(t, "AddIncome", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("no fija ID ni fechas: los deja para que los asigne la base", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(testCategoryID, nil).Once()
		store.On("AddIncome", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.ID == uuid.Nil && i.Fecha.IsZero() && i.CreatedAt.IsZero()
			}),
		).Return(nil).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetCategoryIDByName", mock.Anything, testCategoryName, testEmpresaID).
			Return(testCategoryID, nil).Once()
		store.On("AddIncome", mock.Anything, mock.Anything).
			Return(errStoreBoom).Once()

		_, err := service.AddIncome(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_DeleteIncome(t *testing.T) {
	t.Run("elimina el ingreso con el id recibido", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testEmpresaID).Return(nil).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "DeleteIncome", mock.Anything, testIncomeID, testEmpresaID)
	})

	t.Run("propaga util.ErrNotFound cuando el ingreso no existe", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testEmpresaID).Return(util.ErrNotFound).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("DeleteIncome", mock.Anything, testIncomeID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.DeleteIncome(context.Background(), testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_GetAllIncomesOfEmpresa(t *testing.T) {
	t.Run("devuelve la lista de ingresos de la empresa", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		esperados := []model.Income{newTestIncome()}
		store.On("GetAllIncomesOfEmpresa", mock.Anything, testEmpresaID, newTestFilterQuery()).
			Return(esperados, nil).Once()

		incomes, err := service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.NoError(t, err)
		assert.Equal(t, esperados, incomes)
		store.AssertExpectations(t)
	})

	t.Run("reenvía la paginación al store sin tocarla", func(t *testing.T) {
		// El service es un pasamanos: si normalizara o reinterpretara la
		// paginación, la empresa vería páginas distintas de las que pidió.
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		for _, fq := range []pagination.IncomePaginationQuery{
			filterQueryWithRango(),
			filterQueryWithCategory(),
		} {
			store.On("GetAllIncomesOfEmpresa", mock.Anything, testEmpresaID, fq).
				Return([]model.Income{}, nil).Once()
		}

		_, err := service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, filterQueryWithRango())
		require.NoError(t, err)

		_, err = service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, filterQueryWithCategory())
		require.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("devuelve slice nil sin error cuando la empresa no tiene ingresos", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return(nil, nil).Once()

		incomes, err := service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.NoError(t, err)
		assert.Nil(t, incomes, "sin ingresos debe quedar nil para que el JSON sea null")
		store.AssertExpectations(t)
	})

	t.Run("devuelve nil y el error si el store falla", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return(nil, errStoreBoom).Once()

		incomes, err := service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva incomes", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		store.On("GetAllIncomesOfEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return([]model.Income{newTestIncome()}, errStoreBoom).Once()

		incomes, err := service.GetAllIncomesOfEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes, "si hay error el servicio debe devolver nil")
		store.AssertExpectations(t)
	})

	t.Run("no filtra ingresos de otra empresa: consulta por el ID recibido", func(t *testing.T) {
		store := new(StoreIncomesMock)
		service := NewIncomeService(store)

		otroUsuario := uuid.New()
		store.On("GetAllIncomesOfEmpresa", mock.Anything, otroUsuario, mock.Anything).
			Return(nil, nil).Once()

		_, err := service.GetAllIncomesOfEmpresa(context.Background(), otroUsuario, newTestFilterQuery())

		require.NoError(t, err)
		store.AssertCalled(t, "GetAllIncomesOfEmpresa", mock.Anything, otroUsuario, mock.Anything)
		store.AssertNumberOfCalls(t, "GetAllIncomesOfEmpresa", 1)
	})
}
