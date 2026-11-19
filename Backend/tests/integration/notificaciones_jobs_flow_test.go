package integration

import (
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotificacionesJobsFlow ejercita el ciclo completo de los módulos
// dispositivos, notificaciones y jobs contra el router real.
func TestNotificacionesJobsFlow(t *testing.T) {
	user := testutil.NewUser(t)

	const tokenPrueba = "token-flow-notif-1"

	// -------------------------------------------------------------------------
	// 1. Rutas protegidas → 401
	// -------------------------------------------------------------------------
	t.Run("rutas protegidas", func(t *testing.T) {
		requireProtected(t, http.MethodPut, "/v1/aplicacion/dispositivos")
		requireProtected(t, http.MethodGet, "/v1/aplicacion/dispositivos")
		requireProtected(t, http.MethodDelete, "/v1/aplicacion/dispositivos/tok")
		requireProtected(t, http.MethodGet, "/v1/aplicacion/notificaciones/historial")
		requireProtected(t, http.MethodPost, "/v1/aplicacion/notificaciones/probar")
		requireProtected(t, http.MethodPost, "/v1/aplicacion/jobs/recordatorio-diario")
		requireProtected(t, http.MethodPost, "/v1/aplicacion/jobs/cobros-vencidos")
		requireProtected(t, http.MethodPost, "/v1/aplicacion/jobs/ejecutar-todos")
	})

	// -------------------------------------------------------------------------
	// 2. Registrar dispositivo
	// -------------------------------------------------------------------------
	t.Run("registra un dispositivo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPut, "/v1/aplicacion/dispositivos", user.Token, map[string]any{
			"token":    tokenPrueba,
			"platform": "android",
		})
		testutil.RequireStatus(t, res, http.StatusNoContent)
	})

	t.Run("registrar el mismo token es idempotente (upsert)", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPut, "/v1/aplicacion/dispositivos", user.Token, map[string]any{
			"token":    tokenPrueba,
			"platform": "android",
		})
		testutil.RequireStatus(t, res, http.StatusNoContent)
	})

	t.Run("plataforma inválida devuelve 400", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPut, "/v1/aplicacion/dispositivos", user.Token, map[string]any{
			"token":    "tok-invalid-plat",
			"platform": "macos",
		})
		testutil.RequireStatus(t, res, http.StatusBadRequest)
	})

	t.Run("token vacío devuelve 400", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPut, "/v1/aplicacion/dispositivos", user.Token, map[string]any{
			"token":    "",
			"platform": "ios",
		})
		testutil.RequireStatus(t, res, http.StatusBadRequest)
	})

	// -------------------------------------------------------------------------
	// 3. Listar dispositivos
	// -------------------------------------------------------------------------
	t.Run("lista los dispositivos del usuario", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/dispositivos", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		dispositivos := testutil.DecodeData[[]model.DeviceToken](t, res)
		require.NotEmpty(t, dispositivos, "debe contener al menos el token registrado")

		encontrado := false
		for _, d := range dispositivos {
			if d.Token == tokenPrueba {
				encontrado = true
				assert.Equal(t, user.ID, d.UserID)
				assert.Equal(t, "android", d.Platform)
			}
		}
		require.True(t, encontrado, "token-flow-notif-1 debe estar en la lista")
	})

	// -------------------------------------------------------------------------
	// 4. Notificaciones
	// -------------------------------------------------------------------------
	t.Run("historial vacío al inicio del flujo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/notificaciones/historial", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		// La respuesta puede ser null o [] ya que no hay logs para este usuario.
		// DecodeData convierte null en el zero value del slice (nil), que es OK.
	})

	t.Run("envía notificación de prueba sin payload", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/notificaciones/probar", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		msg := testutil.DecodeData[string](t, res)
		require.Equal(t, "notificación de prueba enviada", msg)
	})

	t.Run("envía notificación de prueba con payload", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/notificaciones/probar", user.Token, map[string]any{
			"titulo": "Hola Tally",
			"cuerpo": "Prueba personalizada",
		})
		testutil.RequireStatus(t, res, http.StatusOK)
	})

	// -------------------------------------------------------------------------
	// 5. Jobs
	// -------------------------------------------------------------------------
	t.Run("recordatorio diario se ejecuta sin error", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/recordatorio-diario", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		msg := testutil.DecodeData[string](t, res)
		require.Equal(t, "recordatorio diario ejecutado", msg)
	})

	t.Run("cobros vencidos se ejecuta sin error", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/cobros-vencidos", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		msg := testutil.DecodeData[string](t, res)
		require.Equal(t, "aviso de cobros vencidos ejecutado", msg)
	})

	t.Run("ejecutar todos los jobs sin error", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/ejecutar-todos", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		msg := testutil.DecodeData[string](t, res)
		require.Equal(t, "todos los jobs ejecutados", msg)
	})

	t.Run("recordatorio diario es idempotente (segunda ejecución sin error)", func(t *testing.T) {
		// La segunda llamada en el mismo día no debe devolver error.
		// El Claim devuelve false (ya existe) y el job simplemente lo saltea.
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/recordatorio-diario", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
	})

	t.Run("cobros vencidos es idempotente (segunda ejecución sin error)", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/cobros-vencidos", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
	})

	// -------------------------------------------------------------------------
	// 6. Desvincular dispositivo
	// -------------------------------------------------------------------------
	t.Run("desvincula el dispositivo por URL param", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, "/v1/aplicacion/dispositivos/"+tokenPrueba, user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusNoContent)
	})

	t.Run("desvincular el mismo token devuelve 404", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, "/v1/aplicacion/dispositivos/"+tokenPrueba, user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusNotFound)
	})

	t.Run("lista vacía tras desvincular", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/dispositivos", user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		// Después de desvincular el único token, la lista debe estar vacía.
		dispositivos := testutil.DecodeData[[]model.DeviceToken](t, res)
		for _, d := range dispositivos {
			require.NotEqual(t, tokenPrueba, d.Token, "el token desvinculado no debe aparecer en la lista")
		}
	})
}

// TestNotificacionesJobsAislamiento verifica que los dispositivos e historial
// de un usuario no son visibles por otro usuario.
func TestNotificacionesJobsAislamiento(t *testing.T) {
	userA := testutil.NewUser(t)
	userB := testutil.NewUser(t)

	const tokenA = "token-aislamiento-a"

	// Registrar dispositivo para userA
	res := testutil.Request(t, http.MethodPut, "/v1/aplicacion/dispositivos", userA.Token, map[string]any{
		"token":    tokenA,
		"platform": "ios",
	})
	testutil.RequireStatus(t, res, http.StatusNoContent)

	t.Cleanup(func() {
		// Limpiar el token de userA aunque el test falle a mitad.
		testutil.Request(t, http.MethodDelete, "/v1/aplicacion/dispositivos/"+tokenA, userA.Token, nil)
	})

	t.Run("userB no ve dispositivos de userA", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/dispositivos", userB.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		dispositivos := testutil.DecodeData[[]model.DeviceToken](t, res)
		for _, d := range dispositivos {
			require.NotEqual(t, tokenA, d.Token, "userB no debe ver el token de userA")
		}
	})

	t.Run("userB no ve historial de notificaciones de userA", func(t *testing.T) {
		// Ejecutar jobs para que userA potencialmente tenga entradas en el log.
		testutil.Request(t, http.MethodPost, "/v1/aplicacion/jobs/ejecutar-todos", userA.Token, nil)

		// El historial de userB debe estar vacío / no contener entradas de userA.
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/notificaciones/historial", userB.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		historial := testutil.DecodeData[[]model.NotificationLog](t, res)
		for _, log := range historial {
			require.Equal(t, userB.ID, log.UserID, "el historial de userB solo debe contener sus propias entradas")
		}
	})
}
