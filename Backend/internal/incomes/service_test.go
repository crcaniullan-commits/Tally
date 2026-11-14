package incomes

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")

func TestIncomeService_AddIncome(t *testing.T) {
	t.Run("mapea el payload a model.Income con el ID de la empresa", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		payload := newTestPayload()

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(true, nil).Once()

		store.On("AgregarIngreso", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.EmpresaID == testEmpresaID &&
					i.Monto == payload.Monto &&
					i.PaymentMethod == utilidades.MetodoPagoDebito &&
					i.Descripcion != nil && *i.Descripcion == "Venta de almuerzo" &&
					i.CategoryID != nil && *i.CategoryID == testCategoryID
			}),
		).Return(nil).Once()

		_, err := service.AgregarIngreso(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertNumberOfCalls(t, "AgregarIngreso", 1)
	})

	t.Run("valida la categoria antes de persistir", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(true, nil).Once()
		store.On("AgregarIngreso", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.CategoryID != nil && *i.CategoryID == testCategoryID
			}),
		).Return(nil).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.NoError(t, err)
		store.AssertCalled(t, "CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID)
		store.AssertExpectations(t)
	})

	t.Run("acepta descripcion nula", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		payload := PayloadIngreso{Monto: 300, PaymentMethod: utilidades.MetodoPagoEfectivo}

		store.On("AgregarIngreso", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool { return i.Descripcion == nil }),
		).Return(nil).Once()

		_, err := service.AgregarIngreso(context.Background(), payload, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("deja la categoria en null si el payload no trae categoria", func(t *testing.T) {
		// incomes.category_id es nullable (migrations/000006). El service no
		// tiene que inventar una categoria por defecto: un nil tiene que
		// seguir siendo nil para que el INSERT escriba NULL y no un UUID cero
		// que no matchearia ninguna fila de categories.
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("AgregarIngreso", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool { return i.CategoryID == nil }),
		).Return(nil).Once()

		_, err := service.AgregarIngreso(context.Background(),
			PayloadIngreso{Monto: 300, PaymentMethod: utilidades.MetodoPagoEfectivo}, testEmpresaID)

		require.NoError(t, err)
		store.AssertNotCalled(t, "CategoriaExiste", mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado si la categoria no existe y no persiste el ingreso", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(false, nil).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertNotCalled(t, "AgregarIngreso", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("rechaza una categoria de otra empresa", func(t *testing.T) {
		// El store responde false porque la categoría existe pero su
		// empresa_id no coincide. El service no debe distinguirla de una
		// inexistente: ambas son 400.
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(false, nil).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertNotCalled(t, "AgregarIngreso", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store de categorias sin persistir", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(false, errStoreBoom).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertNotCalled(t, "AgregarIngreso", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("no fija ID ni fechas: los deja para que los asigne la base", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(true, nil).Once()
		store.On("AgregarIngreso", mock.Anything,
			mock.MatchedBy(func(i *model.Income) bool {
				return i.ID == uuid.Nil && i.Fecha.IsZero() && i.CreatedAt.IsZero()
			}),
		).Return(nil).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store sin envolverlo", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("CategoriaExiste", mock.Anything, testCategoryID, testEmpresaID).
			Return(true, nil).Once()
		store.On("AgregarIngreso", mock.Anything, mock.Anything).
			Return(errStoreBoom).Once()

		_, err := service.AgregarIngreso(context.Background(), newTestPayload(), testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_DeleteIncome(t *testing.T) {
	t.Run("elimina el ingreso con el id recibido", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("EliminarIngreso", mock.Anything, testIncomeID, testEmpresaID).Return(nil).Once()

		err := service.EliminarIngreso(context.Background(), testIncomeID, testEmpresaID)

		require.NoError(t, err)
		store.AssertExpectations(t)
		store.AssertCalled(t, "EliminarIngreso", mock.Anything, testIncomeID, testEmpresaID)
	})

	t.Run("propaga utilidades.ErrorNoEncontrado cuando el ingreso no existe", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("EliminarIngreso", mock.Anything, testIncomeID, testEmpresaID).Return(utilidades.ErrorNoEncontrado).Once()

		err := service.EliminarIngreso(context.Background(), testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertExpectations(t)
	})

	t.Run("propaga errores inesperados del store", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("EliminarIngreso", mock.Anything, testIncomeID, testEmpresaID).Return(errStoreBoom).Once()

		err := service.EliminarIngreso(context.Background(), testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})
}

func TestIncomeService_GetAllIncomesOfEmpresa(t *testing.T) {
	t.Run("devuelve la lista de ingresos de la empresa", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		esperados := []model.Income{newTestIncome()}
		store.On("ObtenerIngresosDeEmpresa", mock.Anything, testEmpresaID, newTestFilterQuery()).
			Return(esperados, nil).Once()

		incomes, err := service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.NoError(t, err)
		assert.Equal(t, esperados, incomes)
		store.AssertExpectations(t)
	})

	t.Run("reenvía la paginación al store sin tocarla", func(t *testing.T) {
		// El service es un pasamanos: si normalizara o reinterpretara la
		// paginación, la empresa vería páginas distintas de las que pidió.
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		for _, fq := range []paginacion.ConsultaPaginacionIngresos{
			filterQueryWithRango(),
			filterQueryWithCategory(),
		} {
			store.On("ObtenerIngresosDeEmpresa", mock.Anything, testEmpresaID, fq).
				Return([]model.Income{}, nil).Once()
		}

		_, err := service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, filterQueryWithRango())
		require.NoError(t, err)

		_, err = service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, filterQueryWithCategory())
		require.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("devuelve slice nil sin error cuando la empresa no tiene ingresos", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("ObtenerIngresosDeEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return(nil, nil).Once()

		incomes, err := service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.NoError(t, err)
		assert.Nil(t, incomes, "sin ingresos debe quedar nil para que el JSON sea null")
		store.AssertExpectations(t)
	})

	t.Run("devuelve nil y el error si el store falla", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("ObtenerIngresosDeEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return(nil, errStoreBoom).Once()

		incomes, err := service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes)
		store.AssertExpectations(t)
	})

	t.Run("no enmascara el error aunque el store devuelva incomes", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		store.On("ObtenerIngresosDeEmpresa", mock.Anything, testEmpresaID, mock.Anything).
			Return([]model.Income{newTestIncome()}, errStoreBoom).Once()

		incomes, err := service.ObtenerIngresosDeEmpresa(context.Background(), testEmpresaID, newTestFilterQuery())

		require.ErrorIs(t, err, errStoreBoom)
		assert.Nil(t, incomes, "si hay error el servicio debe devolver nil")
		store.AssertExpectations(t)
	})

	t.Run("no filtra ingresos de otra empresa: consulta por el ID recibido", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		otroUsuario := uuid.New()
		store.On("ObtenerIngresosDeEmpresa", mock.Anything, otroUsuario, mock.Anything).
			Return(nil, nil).Once()

		_, err := service.ObtenerIngresosDeEmpresa(context.Background(), otroUsuario, newTestFilterQuery())

		require.NoError(t, err)
		store.AssertCalled(t, "ObtenerIngresosDeEmpresa", mock.Anything, otroUsuario, mock.Anything)
		store.AssertNumberOfCalls(t, "ObtenerIngresosDeEmpresa", 1)
	})
}

func TestIncomeService_UpdateIncome(t *testing.T) {
	t.Run("manda al store solo los campos del payload", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		monto := int64(2500)
		descripcion := "Venta de la tarde"
		fecha := "2026-09-15"

		store.On("ActualizarIngreso", mock.Anything,
			mock.MatchedBy(func(c CambiosIngreso) bool {
				return c.Monto != nil && *c.Monto == monto &&
					c.PaymentMethod == nil &&
					c.Descripcion != nil && *c.Descripcion == descripcion &&
					c.CategoryID == nil &&
					c.Fecha != nil && c.Fecha.Format("2006-01-02") == fecha
			}),
			testIncomeID, testEmpresaID,
		).Return(newTestIncome(), nil).Once()

		income, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{
			Monto:       &monto,
			Descripcion: &descripcion,
			Fecha:       &fecha,
		}, testIncomeID, testEmpresaID)

		require.NoError(t, err)
		assert.Equal(t, testIncomeID, income.ID)
		store.AssertExpectations(t)
	})

	t.Run("rechaza una fecha con formato invalido", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		fecha := "15-09-2026"

		_, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{
			Fecha: &fecha,
		}, testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, ErrorFechaInvalida)
		store.AssertNotCalled(t, "ActualizarIngreso", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("rechaza un payload sin campos", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		_, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{}, testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, ErrorSinCambios)
		store.AssertNotCalled(t, "ActualizarIngreso", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("valida la categoria antes de actualizar", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		categoria := uuid.New()

		store.On("CategoriaExiste", mock.Anything, categoria, testEmpresaID).Return(false, nil).Once()

		_, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{
			CategoryID: &categoria,
		}, testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, utilidades.ErrorNoEncontrado)
		store.AssertNotCalled(t, "ActualizarIngreso", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("traduce el ingreso inexistente a ErrorMovimientoNoEncontrado", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		monto := int64(2500)

		store.On("ActualizarIngreso", mock.Anything, mock.Anything, testIncomeID, testEmpresaID).
			Return(model.Income{}, utilidades.ErrorNoEncontrado).Once()

		_, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{
			Monto: &monto,
		}, testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, ErrorMovimientoNoEncontrado)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(RepositorioIngresosMock)
		service := NuevoServicioIngresos(store)

		monto := int64(2500)

		store.On("ActualizarIngreso", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.Income{}, errStoreBoom).Once()

		_, err := service.ActualizarIngreso(context.Background(), PayloadActualizarIngreso{
			Monto: &monto,
		}, testIncomeID, testEmpresaID)

		require.ErrorIs(t, err, errStoreBoom)
	})
}
