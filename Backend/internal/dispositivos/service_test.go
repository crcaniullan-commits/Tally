package dispositivos

import (
	"context"
	"errors"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServicioDispositivos_Registrar(t *testing.T) {
	ctx := context.Background()

	t.Run("falla cuando el token está vacío", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		err := svc.Registrar(ctx, testUserID, PayloadDispositivo{Token: "   ", Platform: "android"})
		assert.ErrorIs(t, err, ErrorTokenRequerido)
		store.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("falla cuando la plataforma no es válida", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		err := svc.Registrar(ctx, testUserID, PayloadDispositivo{Token: testToken, Platform: "windows"})
		assert.ErrorIs(t, err, ErrorPlataformaInvalida)
		store.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("registra correctamente con plataforma normalizada", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		store.On("Upsert", ctx, testUserID, testToken, "ios").Return(nil).Once()

		err := svc.Registrar(ctx, testUserID, PayloadDispositivo{Token: testToken, Platform: "  IOS  "})
		assert.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("propaga error del store al registrar", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		expectedErr := errors.New("db error")
		store.On("Upsert", ctx, testUserID, testToken, "android").Return(expectedErr).Once()

		err := svc.Registrar(ctx, testUserID, PayloadDispositivo{Token: testToken, Platform: "android"})
		assert.ErrorIs(t, err, expectedErr)
		store.AssertExpectations(t)
	})
}

func TestServicioDispositivos_Desvincular(t *testing.T) {
	ctx := context.Background()

	t.Run("falla si el token está vacío", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		err := svc.Desvincular(ctx, testUserID, "   ")
		assert.ErrorIs(t, err, ErrorTokenRequerido)
		store.AssertNotCalled(t, "Eliminar", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("desvincula correctamente", func(t *testing.T) {
		store := new(DispositivosStoreMock)
		svc := NuevoServicioDispositivos(store)

		store.On("Eliminar", ctx, testUserID, testToken).Return(nil).Once()

		err := svc.Desvincular(ctx, testUserID, testToken)
		assert.NoError(t, err)
		store.AssertExpectations(t)
	})
}

func TestServicioDispositivos_Listar(t *testing.T) {
	ctx := context.Background()
	store := new(DispositivosStoreMock)
	svc := NuevoServicioDispositivos(store)

	expected := []model.DeviceToken{
		{
			ID:         testDeviceID,
			UserID:     testUserID,
			Token:      testToken,
			Platform:   "android",
			LastSeenAt: testNow,
		},
	}

	store.On("ListarPorUsuario", ctx, testUserID).Return(expected, nil).Once()

	dispositivos, err := svc.Listar(ctx, testUserID)
	assert.NoError(t, err)
	assert.Equal(t, expected, dispositivos)
	store.AssertExpectations(t)
}

func TestServicioDispositivos_ObtenerTokensYEliminar(t *testing.T) {
	ctx := context.Background()
	store := new(DispositivosStoreMock)
	svc := NuevoServicioDispositivos(store)

	tokens := []string{testToken}
	store.On("ObtenerTokensPorUsuario", ctx, testUserID).Return(tokens, nil).Once()
	res, err := svc.ObtenerTokensPorUsuario(ctx, testUserID)
	assert.NoError(t, err)
	assert.Equal(t, tokens, res)

	store.On("EliminarTokens", ctx, tokens).Return(nil).Once()
	err = svc.EliminarTokensInvalidos(ctx, tokens)
	assert.NoError(t, err)

	store.AssertExpectations(t)
}
