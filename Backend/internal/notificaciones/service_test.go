package notificaciones

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestServicioNotificaciones_Send(t *testing.T) {
	ctx := context.Background()
	n := model.Notification{
		Title: "Título de prueba",
		Body:  "Cuerpo de prueba",
		Data:  map[string]string{"pantalla": "home"},
	}

	t.Run("no hace nada si dispositivos o transporte son nil", func(t *testing.T) {
		svc := NuevoServicioNotificaciones(nil, nil, nil, nil)
		err := svc.Send(ctx, testUserID, n)
		assert.NoError(t, err)
	})

	t.Run("propaga error si falla obtener tokens", func(t *testing.T) {
		store := new(NotificacionesStoreMock)
		disp := new(BuscadorDispositivosMock)
		push := new(EnviadorPushMock)
		svc := NuevoServicioNotificaciones(store, push, disp, zap.NewNop().Sugar())

		errMock := errors.New("error tokens")
		disp.On("ObtenerTokensPorUsuario", ctx, testUserID).Return(nil, errMock).Once()

		err := svc.Send(ctx, testUserID, n)
		assert.ErrorIs(t, err, errMock)
		push.AssertNotCalled(t, "EnviarPush", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("retorna nil si el usuario no tiene dispositivos registrados", func(t *testing.T) {
		store := new(NotificacionesStoreMock)
		disp := new(BuscadorDispositivosMock)
		push := new(EnviadorPushMock)
		svc := NuevoServicioNotificaciones(store, push, disp, zap.NewNop().Sugar())

		disp.On("ObtenerTokensPorUsuario", ctx, testUserID).Return([]string{}, nil).Once()

		err := svc.Send(ctx, testUserID, n)
		assert.NoError(t, err)
		push.AssertNotCalled(t, "EnviarPush", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("envía push exitosamente sin tokens inválidos", func(t *testing.T) {
		store := new(NotificacionesStoreMock)
		disp := new(BuscadorDispositivosMock)
		push := new(EnviadorPushMock)
		svc := NuevoServicioNotificaciones(store, push, disp, zap.NewNop().Sugar())

		disp.On("ObtenerTokensPorUsuario", ctx, testUserID).Return(testTokens, nil).Once()
		push.On("EnviarPush", ctx, testTokens, n).Return([]string{}, nil).Once()

		err := svc.Send(ctx, testUserID, n)
		assert.NoError(t, err)
		disp.AssertNotCalled(t, "EliminarTokensInvalidos", mock.Anything, mock.Anything)
		push.AssertExpectations(t)
	})

	t.Run("elimina tokens si el transporte reporta tokens inválidos", func(t *testing.T) {
		store := new(NotificacionesStoreMock)
		disp := new(BuscadorDispositivosMock)
		push := new(EnviadorPushMock)
		svc := NuevoServicioNotificaciones(store, push, disp, zap.NewNop().Sugar())

		disp.On("ObtenerTokensPorUsuario", ctx, testUserID).Return(testTokens, nil).Once()
		push.On("EnviarPush", ctx, testTokens, n).Return(testTokens, nil).Once()
		disp.On("EliminarTokensInvalidos", ctx, testTokens).Return(nil).Once()

		err := svc.Send(ctx, testUserID, n)
		assert.NoError(t, err)
		disp.AssertExpectations(t)
		push.AssertExpectations(t)
	})

	t.Run("propaga error si el transporte push falla", func(t *testing.T) {
		store := new(NotificacionesStoreMock)
		disp := new(BuscadorDispositivosMock)
		push := new(EnviadorPushMock)
		svc := NuevoServicioNotificaciones(store, push, disp, zap.NewNop().Sugar())

		errPush := errors.New("fallo expo")
		disp.On("ObtenerTokensPorUsuario", ctx, testUserID).Return(testTokens, nil).Once()
		push.On("EnviarPush", ctx, testTokens, n).Return(nil, errPush).Once()

		err := svc.Send(ctx, testUserID, n)
		assert.ErrorIs(t, err, errPush)
	})
}

func TestServicioNotificaciones_ClaimUnclaimListar(t *testing.T) {
	ctx := context.Background()
	store := new(NotificacionesStoreMock)
	svc := NuevoServicioNotificaciones(store, nil, nil, nil)

	store.On("Claim", ctx, testUserID, "daily", "2026-11-18").Return(true, nil).Once()
	ok, err := svc.Claim(ctx, testUserID, "daily", "2026-11-18")
	assert.NoError(t, err)
	assert.True(t, ok)

	store.On("Unclaim", ctx, testUserID, "daily", "2026-11-18").Return(nil).Once()
	err = svc.Unclaim(ctx, testUserID, "daily", "2026-11-18")
	assert.NoError(t, err)

	expectedLogs := []model.NotificationLog{
		{UserID: testUserID, Kind: "daily"},
	}
	store.On("ListarPorUsuario", ctx, testUserID).Return(expectedLogs, nil).Once()
	logs, err := svc.Listar(ctx, testUserID)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)

	store.AssertExpectations(t)
}
