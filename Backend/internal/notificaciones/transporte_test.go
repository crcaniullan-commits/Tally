package notificaciones

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogNotifier(t *testing.T) {
	notifier := NuevoLogNotifier(zap.NewNop().Sugar())
	invalid, err := notifier.EnviarPush(context.Background(), []string{"token1"}, model.Notification{
		Title: "Alerta",
		Body:  "Mensaje de prueba",
	})
	assert.NoError(t, err)
	assert.Nil(t, invalid)
}

func TestClienteExpoPush(t *testing.T) {
	ctx := context.Background()
	n := model.Notification{
		Title: "Cobros vencidos",
		Body:  "Tienes cobros pendientes",
	}

	t.Run("tokens vacío retorna nil inmediatamente", func(t *testing.T) {
		cliente := NuevoClienteExpoPush(nil)
		invalidos, err := cliente.EnviarPush(ctx, nil, n)
		assert.NoError(t, err)
		assert.Nil(t, invalidos)
	})

	t.Run("envío exitoso con status ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{"status": "ok", "id": "1234"}
				]
			}`))
		}))
		defer ts.Close()

		cliente := NuevoClienteExpoPush(ts.Client()).ConEndpoint(ts.URL)
		invalidos, err := cliente.EnviarPush(ctx, []string{"token-valido"}, n)
		assert.NoError(t, err)
		assert.Empty(t, invalidos)
	})

	t.Run("detecta token inválido DeviceNotRegistered", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"status": "error",
						"message": "Device not registered",
						"details": {"error": "DeviceNotRegistered"}
					}
				]
			}`))
		}))
		defer ts.Close()

		cliente := NuevoClienteExpoPush(ts.Client()).ConEndpoint(ts.URL)
		invalidos, err := cliente.EnviarPush(ctx, []string{"token-expirado"}, n)
		assert.NoError(t, err)
		assert.Equal(t, []string{"token-expirado"}, invalidos)
	})

	t.Run("retorna error si el servidor responde 500", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		cliente := NuevoClienteExpoPush(ts.Client()).ConEndpoint(ts.URL)
		invalidos, err := cliente.EnviarPush(ctx, []string{"token"}, n)
		assert.Error(t, err)
		assert.Nil(t, invalidos)
	})
}
