package integration

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEmpresasFlow(t *testing.T) {
	user := testutil.NewUser(t)
	name := testutil.UniqueName("Empresa Flow")

	var empresa model.Empresa

	t.Run("crea una empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/empresas", user.Token, map[string]string{
			"nombre_empresa": name,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		empresa = testutil.DecodeData[model.Empresa](t, res)
		require.NotEqual(t, uuid.Nil, empresa.ID)
		require.Equal(t, name, empresa.NombreEmpresa)
		require.Equal(t, user.ID, empresa.UserID)
	})

	t.Run("lista las empresas del usuario", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		empresas := testutil.DecodeData[[]model.Empresa](t, res)
		require.Contains(t, empresas, empresa)
	})

	t.Run("renombra la empresa", func(t *testing.T) {
		newName := testutil.UniqueName("Empresa Renombrada")
		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/"+url.PathEscape(newName)), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Nombre actualizado", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", user.Token, nil)
		empresas := testutil.DecodeData[[]model.Empresa](t, list)
		require.Contains(t, empresas, model.Empresa{ID: empresa.ID, NombreEmpresa: newName, UserID: user.ID})
	})

	t.Run("el rename no valida el largo del nombre", func(t *testing.T) {
		// DISCREPANCIA: el payload de creación limita nombre_empresa a 30
		// caracteres, pero PATCH /empresas/{id}/{nombre} toma el nombre del
		// path y no lo valida. Un nombre de 31 caracteres se acepta.
		longName := "1234567890123456789012345678901"
		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/"+longName), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Nombre actualizado", testutil.DecodeData[string](t, res))
	})

	t.Run("borra la empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, ""), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "empresa borrada", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", user.Token, nil)
		empresas := testutil.DecodeData[[]model.Empresa](t, list)
		for _, e := range empresas {
			require.NotEqual(t, empresa.ID, e.ID)
		}
	})

	t.Run("borrar de nuevo devuelve 404", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, ""), user.Token, nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("un usuario sin empresas recibe data null", func(t *testing.T) {
		// DISCREPANCIA: el swagger documenta 404 para GET /aplicacion/empresas cuando
		// no hay empresas; el handler responde 200 con data null.
		empty := testutil.NewUser(t)
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", empty.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Nil(t, testutil.DecodeData[[]model.Empresa](t, res))
	})

	t.Run("creación valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "nombre faltante",
				body:    map[string]any{},
				message: "Key: 'PayloadEmpresa.NombreEmpresa' Error:Field validation for 'NombreEmpresa' failed on the 'required' tag",
			},
			{
				name:    "nombre demasiado largo",
				body:    map[string]string{"nombre_empresa": "1234567890123456789012345678901"},
				message: "Key: 'PayloadEmpresa.NombreEmpresa' Error:Field validation for 'NombreEmpresa' failed on the 'max' tag",
			},
			{
				name:    "tipo incorrecto",
				body:    `{"nombre_empresa":123}`,
				message: "json: cannot unmarshal number into Go struct field PayloadEmpresa.nombre_empresa of type string",
			},
			{
				name:    "campo desconocido",
				body:    map[string]any{"nombre_empresa": "X", "extra": 1},
				message: `json: unknown field "extra"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/empresas", user.Token, tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("un municipal no puede crear empresas", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/aplicacion/empresas", testutil.MunicipalToken(t), map[string]string{
			"nombre_empresa": "Empresa Municipal",
		})
		testutil.RequireError(t, res, http.StatusForbidden, "prohibido")
	})

	t.Run("un municipal no puede listar empresas", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusForbidden, "prohibido")
	})

	t.Run("no puede tocar la empresa de otro usuario", func(t *testing.T) {
		owner := testutil.NewUser(t)
		owned := testutil.CreateEmpresa(t, owner)
		intruder := testutil.NewUser(t)

		patch := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(owned.ID, "/Intruso"), intruder.Token, nil)
		testutil.RequireError(t, patch, http.StatusNotFound, "no encontrado")

		del := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(owned.ID, ""), intruder.Token, nil)
		testutil.RequireError(t, del, http.StatusNotFound, "no encontrado")
	})

	t.Run("errores de ruta", func(t *testing.T) {
		cases := []struct {
			name    string
			method  string
			path    string
			status  int
			message string
		}{
			{
				name:    "delete con uuid inválido",
				method:  http.MethodDelete,
				path:    "/v1/aplicacion/empresas/no-uuid",
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "patch con uuid inválido",
				method:  http.MethodPatch,
				path:    "/v1/aplicacion/empresas/no-uuid/Nueva",
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "delete inexistente",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(uuid.New(), ""),
				status:  http.StatusNotFound,
				message: "no encontrado",
			},
			{
				name:    "patch inexistente",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(uuid.New(), "/Nueva"),
				status:  http.StatusNotFound,
				message: "no encontrado",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, tc.method, tc.path, user.Token, nil)
				testutil.RequireError(t, res, tc.status, tc.message)
			})
		}
	})

	requireProtected(t, http.MethodGet, "/v1/aplicacion/empresas")
	requireProtected(t, http.MethodPost, "/v1/aplicacion/empresas")
	requireProtected(t, http.MethodDelete, testutil.EmpresaPath(uuid.New(), ""))
}
