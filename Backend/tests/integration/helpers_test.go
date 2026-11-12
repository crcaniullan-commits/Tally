package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/require"
)

// requireProtected verifica que una ruta detrás de MiddlewareTokenAutenticacion
// responda 401 tanto sin header Authorization como con un token inválido.
func requireProtected(t *testing.T, method, path string) {
	t.Helper()

	t.Run("sin token responde 401", func(t *testing.T) {
		res := testutil.Request(t, method, path, "", nil)
		testutil.RequireError(t, res, http.StatusUnauthorized, "no autorizado")
	})

	t.Run("token inválido responde 401", func(t *testing.T) {
		res := testutil.Request(t, method, path, "not-a-token", nil)
		testutil.RequireError(t, res, http.StatusUnauthorized, "no autorizado")
	})
}

// dbDate devuelve CURRENT_DATE según la base, que es la que usa la columna
// incomes.fecha. Evita diferencias de zona horaria con el reloj del test.
func dbDate(t *testing.T, offsetDays int) string {
	t.Helper()

	var current time.Time
	require.NoError(t, testutil.GetEnv().DB.QueryRow(`SELECT CURRENT_DATE`).Scan(&current))

	return current.AddDate(0, 0, offsetDays).Format(time.DateOnly)
}
