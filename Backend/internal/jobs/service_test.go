package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestFormatearCLP(t *testing.T) {
	assert.Equal(t, "$0", FormatearCLP(0))
	assert.Equal(t, "$500", FormatearCLP(500))
	assert.Equal(t, "$1.000", FormatearCLP(1000))
	assert.Equal(t, "$45.000", FormatearCLP(45000))
	assert.Equal(t, "$1.500.500", FormatearCLP(1500500))
}

func TestServicioJobs_RecordatorioDiario(t *testing.T) {
	ctx := context.Background()
	loc, _ := time.LoadLocation("America/Santiago")

	t.Run("propaga error si el store falla", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		errStore := errors.New("fallo store")
		store.On("UsuariosSinMovimientos", ctx, mock.Anything, mock.Anything).
			Return(nil, errStore).Once()

		err := svc.RecordatorioDiario(ctx)
		assert.ErrorIs(t, err, errStore)
		notif.AssertNotCalled(t, "Claim", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("envía notificación cuando obtiene candado", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		u1 := uuid.New()
		u2 := uuid.New()
		store.On("UsuariosSinMovimientos", ctx, mock.Anything, mock.Anything).
			Return([]uuid.UUID{u1, u2}, nil).Once()

		// u1 obtiene candado y envía exitosamente
		notif.On("Claim", ctx, u1, "daily", mock.Anything).Return(true, nil).Once()
		notif.On("Send", ctx, u1, mock.MatchedBy(func(n model.Notification) bool {
			return n.Title == "¿Cómo va tu día?" && n.Data["screen"] == "incomes"
		})).Return(nil).Once()

		// u2 ya fue notificado (candado false)
		notif.On("Claim", ctx, u2, "daily", mock.Anything).Return(false, nil).Once()

		err := svc.RecordatorioDiario(ctx)
		assert.NoError(t, err)

		notif.AssertNotCalled(t, "Send", ctx, u2, mock.Anything)
		notif.AssertNotCalled(t, "Unclaim", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		store.AssertExpectations(t)
		notif.AssertExpectations(t)
	})

	t.Run("hace unclaim si el envío push falla", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		u := uuid.New()
		store.On("UsuariosSinMovimientos", ctx, mock.Anything, mock.Anything).
			Return([]uuid.UUID{u}, nil).Once()

		notif.On("Claim", ctx, u, "daily", mock.Anything).Return(true, nil).Once()
		notif.On("Send", ctx, u, mock.Anything).Return(errors.New("fallo push")).Once()
		notif.On("Unclaim", ctx, u, "daily", mock.Anything).Return(nil).Once()

		err := svc.RecordatorioDiario(ctx)
		assert.NoError(t, err) // El job atrapa el error por usuario para no abortar el ciclo

		store.AssertExpectations(t)
		notif.AssertExpectations(t)
	})
}

func TestServicioJobs_CobrosVencidos(t *testing.T) {
	ctx := context.Background()
	loc, _ := time.LoadLocation("America/Santiago")

	t.Run("propaga error si el store falla", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		errStore := errors.New("fallo store")
		store.On("CobrosVencidosPorUsuario", ctx, mock.Anything).
			Return(nil, errStore).Once()

		err := svc.CobrosVencidos(ctx)
		assert.ErrorIs(t, err, errStore)
	})

	t.Run("envía notificación con formato singular y plural", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		u1 := uuid.New()
		u2 := uuid.New()
		resumenes := []ResumenCobrosVencidos{
			{UserID: u1, Cantidad: 1, Total: 10000},
			{UserID: u2, Cantidad: 3, Total: 45000},
		}

		store.On("CobrosVencidosPorUsuario", ctx, mock.Anything).
			Return(resumenes, nil).Once()

		// u1: singular
		notif.On("Claim", ctx, u1, "overdue", mock.Anything).Return(true, nil).Once()
		notif.On("Send", ctx, u1, mock.MatchedBy(func(n model.Notification) bool {
			return n.Body == "Tienes 1 cobro vencido por $10.000." && n.Data["screen"] == "debtors"
		})).Return(nil).Once()

		// u2: plural
		notif.On("Claim", ctx, u2, "overdue", mock.Anything).Return(true, nil).Once()
		notif.On("Send", ctx, u2, mock.MatchedBy(func(n model.Notification) bool {
			return n.Body == "Tienes 3 cobros vencidos por $45.000." && n.Data["screen"] == "debtors"
		})).Return(nil).Once()

		err := svc.CobrosVencidos(ctx)
		assert.NoError(t, err)

		store.AssertExpectations(t)
		notif.AssertExpectations(t)
	})

	t.Run("hace unclaim si el envío push falla", func(t *testing.T) {
		store := new(JobsStoreMock)
		notif := new(NotificadorMock)
		svc := NuevoServicioJobs(store, notif, loc, zap.NewNop().Sugar())

		u := uuid.New()
		store.On("CobrosVencidosPorUsuario", ctx, mock.Anything).
			Return([]ResumenCobrosVencidos{{UserID: u, Cantidad: 1, Total: 5000}}, nil).Once()

		notif.On("Claim", ctx, u, "overdue", mock.Anything).Return(true, nil).Once()
		notif.On("Send", ctx, u, mock.Anything).Return(errors.New("fallo push")).Once()
		notif.On("Unclaim", ctx, u, "overdue", mock.Anything).Return(nil).Once()

		err := svc.CobrosVencidos(ctx)
		assert.NoError(t, err)

		store.AssertExpectations(t)
		notif.AssertExpectations(t)
	})
}

func TestServicioJobs_EjecutarTodos(t *testing.T) {
	ctx := context.Background()
	store := new(JobsStoreMock)
	notif := new(NotificadorMock)
	svc := NuevoServicioJobs(store, notif, nil, nil)

	store.On("UsuariosSinMovimientos", ctx, mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil).Once()
	store.On("CobrosVencidosPorUsuario", ctx, mock.Anything).Return([]ResumenCobrosVencidos{}, nil).Once()

	err := svc.EjecutarTodos(ctx)
	assert.NoError(t, err)

	store.AssertExpectations(t)
}
