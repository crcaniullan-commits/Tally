package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDebtorsFlow(t *testing.T) {
	user := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, user)

	nombre := testutil.UniqueName("Deudor Flow")
	fechaLimite := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

	var debtor model.Debtor

	t.Run("crea un deudor", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, map[string]any{
			"nombre":       nombre,
			"monto":        50000,
			"descripcion":  "deuda de prueba",
			"fecha_limite": fechaLimite.Format(time.RFC3339),
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		debtor = testutil.DecodeData[model.Debtor](t, res)
		require.NotEqual(t, uuid.Nil, debtor.ID)
		require.Equal(t, empresa.ID, debtor.EmpresaID)
		require.Equal(t, nombre, debtor.Nombre)
		require.Equal(t, int64(50000), debtor.Monto)
		require.NotNil(t, debtor.Descripcion)
		require.Equal(t, "deuda de prueba", *debtor.Descripcion)
		require.False(t, debtor.Pagado)
		require.NotNil(t, debtor.FechaLimite)
		require.True(t, debtor.FechaLimite.Equal(fechaLimite))
	})

	t.Run("lista los deudores de la empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		debtors := testutil.DecodeData[[]model.Debtor](t, res)

		found := false
		for _, d := range debtors {
			if d.ID == debtor.ID {
				found = true
			}
		}
		require.True(t, found)
	})

	t.Run("actualiza nombre, monto y fecha límite", func(t *testing.T) {
		newName := testutil.UniqueName("Deudor Actualizado")
		newDate := time.Now().Add(72 * time.Hour).UTC().Truncate(time.Second)

		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+debtor.ID.String()), user.Token, map[string]any{
			"nombre":       newName,
			"monto":        60000,
			"fecha_limite": newDate.Format(time.RFC3339),
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Deuda actualizada", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, nil)
		debtors := testutil.DecodeData[[]model.Debtor](t, list)
		for _, d := range debtors {
			if d.ID == debtor.ID {
				require.Equal(t, newName, d.Nombre)
				require.Equal(t, int64(60000), d.Monto)
				require.NotNil(t, d.FechaLimite)
				require.Equal(t, newDate.Format(time.DateOnly), d.FechaLimite.Format(time.DateOnly))
				// El update no toca la descripción.
				require.NotNil(t, d.Descripcion)
				require.Equal(t, "deuda de prueba", *d.Descripcion)
			}
		}
	})

	t.Run("marca el deudor como pagado", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+debtor.ID.String()+"/pagada"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Deuda pagada", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, nil)
		debtors := testutil.DecodeData[[]model.Debtor](t, list)
		for _, d := range debtors {
			if d.ID == debtor.ID {
				require.True(t, d.Pagado)
			}
		}
	})

	t.Run("creación valida el payload", func(t *testing.T) {
		valid := map[string]any{
			"nombre":       testutil.UniqueName("Válido"),
			"monto":        1000,
			"descripcion":  "d",
			"fecha_limite": time.Now().UTC().Format(time.RFC3339),
		}

		cases := []struct {
			name    string
			mutate  func(map[string]any)
			message string
		}{
			{
				name:    "nombre faltante",
				mutate:  func(m map[string]any) { delete(m, "nombre") },
				message: "Key: 'PayloadDeudor.Nombre' Error:Field validation for 'Nombre' failed on the 'required' tag",
			},
			{
				name:    "monto faltante",
				mutate:  func(m map[string]any) { delete(m, "monto") },
				message: "Key: 'PayloadDeudor.Monto' Error:Field validation for 'Monto' failed on the 'required' tag",
			},
			{
				name:    "monto cero",
				mutate:  func(m map[string]any) { m["monto"] = 0 },
				message: "Key: 'PayloadDeudor.Monto' Error:Field validation for 'Monto' failed on the 'required' tag",
			},
			{
				name:    "monto negativo",
				mutate:  func(m map[string]any) { m["monto"] = -1 },
				message: "Key: 'PayloadDeudor.Monto' Error:Field validation for 'Monto' failed on the 'gt' tag",
			},
			{
				name:    "descripción faltante",
				mutate:  func(m map[string]any) { delete(m, "descripcion") },
				message: "Key: 'PayloadDeudor.Descripcion' Error:Field validation for 'Descripcion' failed on the 'required' tag",
			},
			{
				name:    "fecha límite faltante",
				mutate:  func(m map[string]any) { delete(m, "fecha_limite") },
				message: "Key: 'PayloadDeudor.FechaLimite' Error:Field validation for 'FechaLimite' failed on the 'required' tag",
			},
			{
				name:    "nombre demasiado largo",
				mutate:  func(m map[string]any) { m["nombre"] = "1234567890123456789012345678901" },
				message: "Key: 'PayloadDeudor.Nombre' Error:Field validation for 'Nombre' failed on the 'max' tag",
			},
			{
				name:    "descripción demasiado larga",
				mutate:  func(m map[string]any) { m["descripcion"] = string(make([]byte, 101)) },
				message: "Key: 'PayloadDeudor.Descripcion' Error:Field validation for 'Descripcion' failed on the 'max' tag",
			},
			{
				name:    "fecha límite con formato inválido",
				mutate:  func(m map[string]any) { m["fecha_limite"] = "2026-12-31" },
				message: `parsing time "2026-12-31" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				body := map[string]any{}
				for k, v := range valid {
					body[k] = v
				}
				tc.mutate(body)

				res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("rechaza campos desconocidos al crear", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/deudas"), user.Token, map[string]any{
			"nombre":       "X",
			"monto":        1000,
			"descripcion":  "d",
			"fecha_limite": time.Now().UTC().Format(time.RFC3339),
			"extra":        1,
		})
		testutil.RequireError(t, res, http.StatusBadRequest, `json: unknown field "extra"`)
	})

	t.Run("actualización valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "nombre faltante",
				body:    map[string]any{"monto": 1000, "fecha_limite": time.Now().UTC().Format(time.RFC3339)},
				message: "Key: 'PayloadActualizacionDeudor.Nombre' Error:Field validation for 'Nombre' failed on the 'required' tag",
			},
			{
				name:    "monto faltante",
				body:    map[string]any{"nombre": "X", "fecha_limite": time.Now().UTC().Format(time.RFC3339)},
				message: "Key: 'PayloadActualizacionDeudor.Monto' Error:Field validation for 'Monto' failed on the 'required' tag",
			},
			{
				name:    "fecha faltante",
				body:    map[string]any{"nombre": "X", "monto": 1000},
				message: "Key: 'PayloadActualizacionDeudor.FechaLimite' Error:Field validation for 'FechaLimite' failed on the 'required' tag",
			},
			{
				name:    "descripción no permitida",
				body:    map[string]any{"nombre": "X", "monto": 1000, "fecha_limite": time.Now().UTC().Format(time.RFC3339), "descripcion": "d"},
				message: `json: unknown field "descripcion"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+debtor.ID.String()), user.Token, tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("errores de ruta", func(t *testing.T) {
		validBody := map[string]any{
			"nombre":       "X",
			"monto":        1000,
			"fecha_limite": time.Now().UTC().Format(time.RFC3339),
		}

		cases := []struct {
			name    string
			method  string
			path    string
			body    any
			status  int
			message string
		}{
			{
				name:    "patch con uuid inválido",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(empresa.ID, "/deudas/no-uuid"),
				body:    validBody,
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "patch inexistente",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(empresa.ID, "/deudas/"+uuid.NewString()),
				body:    validBody,
				status:  http.StatusNotFound,
				message: "no encontrado",
			},
			{
				name:    "paid con uuid inválido",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(empresa.ID, "/deudas/no-uuid/pagada"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "paid inexistente",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(empresa.ID, "/deudas/"+uuid.NewString()+"/pagada"),
				status:  http.StatusNotFound,
				message: "no encontrado",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, tc.method, tc.path, user.Token, tc.body)
				testutil.RequireError(t, res, tc.status, tc.message)
			})
		}
	})

	t.Run("no puede tocar deudores de otra empresa", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)
		otherDebtor := testutil.CreateDebtor(t, otherUser, otherEmpresa.ID, 1000)

		patch := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+otherDebtor.ID.String()), user.Token, map[string]any{
			"nombre":       "Intruso",
			"monto":        2000,
			"fecha_limite": time.Now().UTC().Format(time.RFC3339),
		})
		testutil.RequireError(t, patch, http.StatusNotFound, "no encontrado")

		paid := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+otherDebtor.ID.String()+"/pagada"), user.Token, nil)
		testutil.RequireError(t, paid, http.StatusNotFound, "no encontrado")
	})

	t.Run("no puede acceder a la empresa de otro usuario", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/deudas"), user.Token, nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("un municipal no puede acceder a los deudores", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/deudas"), testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/deudas"))
	requireProtected(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/deudas"))
	requireProtected(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+uuid.NewString()))
	requireProtected(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/deudas/"+uuid.NewString()+"/pagada"))
}
