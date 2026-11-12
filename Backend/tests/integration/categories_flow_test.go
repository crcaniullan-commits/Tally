package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCategoriesFlow(t *testing.T) {
	user := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, user)

	var categoria model.Categorie

	t.Run("crea una categoría propia", func(t *testing.T) {
		name := testutil.UniqueName("Cat Flow")
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": name,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		categoria = testutil.DecodeData[model.Categorie](t, res)
		require.NotEqual(t, uuid.Nil, categoria.ID)
		require.Equal(t, name, categoria.Nombre)
		require.NotNil(t, categoria.EmpresaID)
		require.Equal(t, empresa.ID, *categoria.EmpresaID)
	})

	t.Run("lista las categorías propias y las globales", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		categorias := testutil.DecodeData[[]model.Categorie](t, res)

		ownIndex, globalIndex := -1, -1
		for i, c := range categorias {
			if c.ID == categoria.ID {
				ownIndex = i
			}
			if c.EmpresaID == nil && globalIndex == -1 {
				globalIndex = i
			}
		}

		require.NotEqual(t, -1, ownIndex, "la categoría propia debería aparecer")
		require.NotEqual(t, -1, globalIndex, "deberían aparecer las categorías globales")
		require.Less(t, ownIndex, globalIndex, "las categorías propias van antes que las globales")

		for _, nombre := range []string{"Ventas", "Servicios", "Pedidos"} {
			found := false
			for _, c := range categorias {
				if c.EmpresaID == nil && c.Nombre == nombre {
					found = true
					break
				}
			}
			require.True(t, found, "debería estar la categoría global %q", nombre)
		}
	})

	t.Run("renombra la categoría", func(t *testing.T) {
		newName := testutil.UniqueName("Cat Renombrada")
		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/categorias/"+categoria.ID.String()), user.Token, map[string]string{
			"nombre": newName,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Nombre cambiado", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, nil)
		categorias := testutil.DecodeData[[]model.Categorie](t, list)
		for _, c := range categorias {
			if c.ID == categoria.ID {
				require.Equal(t, newName, c.Nombre)
			}
		}
	})

	t.Run("rechaza duplicados ignorando mayúsculas", func(t *testing.T) {
		base := testutil.UniqueName("Comida")
		first := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": base,
		})
		testutil.RequireStatus(t, first, http.StatusCreated)

		second := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": strings.ToUpper(base),
		})
		testutil.RequireError(t, second, http.StatusBadRequest, "La categoría ya existe")
	})

	t.Run("rechaza duplicados ignorando acentos", func(t *testing.T) {
		suffix := uuid.NewString()[:8]
		first := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": "Áqui " + suffix,
		})
		testutil.RequireStatus(t, first, http.StatusCreated)

		second := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": "aqui " + suffix,
		})
		testutil.RequireError(t, second, http.StatusBadRequest, "La categoría ya existe")
	})

	t.Run("rechaza renombrar a un nombre ya usado", func(t *testing.T) {
		base := testutil.UniqueName("Repetida")
		created := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": base,
		})
		testutil.RequireStatus(t, created, http.StatusCreated)

		other := testutil.CreateCategory(t, user, empresa.ID)
		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/categorias/"+other.ID.String()), user.Token, map[string]string{
			"nombre": strings.ToLower(base),
		})
		testutil.RequireError(t, res, http.StatusBadRequest, "La categoría ya existe")
	})

	t.Run("el mismo nombre puede existir en otra empresa", func(t *testing.T) {
		name := testutil.UniqueName("Compartida")
		first := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": name,
		})
		testutil.RequireStatus(t, first, http.StatusCreated)

		otherEmpresa := testutil.CreateEmpresa(t, user)
		second := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(otherEmpresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": name,
		})
		testutil.RequireStatus(t, second, http.StatusCreated)
	})

	t.Run("permite repetir el nombre de una categoría global", func(t *testing.T) {
		// La unicidad solo se valida contra las categorías de la empresa; el
		// nombre de una global no bloquea una categoría propia.
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, map[string]string{
			"nombre": "Ventas",
		})
		testutil.RequireStatus(t, res, http.StatusCreated)
	})

	t.Run("no se pueden modificar ni borrar categorías globales desde la empresa", func(t *testing.T) {
		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, nil)
		categorias := testutil.DecodeData[[]model.Categorie](t, list)

		var global model.Categorie
		for _, c := range categorias {
			if c.EmpresaID == nil {
				global = c
				break
			}
		}
		require.NotEqual(t, uuid.Nil, global.ID)

		patch := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/categorias/"+global.ID.String()), user.Token, map[string]string{
			"nombre": "Global Renombrada",
		})
		testutil.RequireError(t, patch, http.StatusNotFound, "no encontrado")

		del := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/categorias/"+global.ID.String()), user.Token, nil)
		testutil.RequireError(t, del, http.StatusNotFound, "no encontrado")
	})

	t.Run("borra la categoría", func(t *testing.T) {
		toDelete := testutil.CreateCategory(t, user, empresa.ID)
		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/categorias/"+toDelete.ID.String()), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, fmt.Sprintf("Categoría %v eliminada", toDelete.ID), testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, nil)
		categorias := testutil.DecodeData[[]model.Categorie](t, list)
		for _, c := range categorias {
			require.NotEqual(t, toDelete.ID, c.ID)
		}
	})

	t.Run("creación valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "nombre vacío",
				body:    map[string]string{"nombre": ""},
				message: "Key: 'PayloadCategoria.Nombre' Error:Field validation for 'Nombre' failed on the 'required' tag",
			},
			{
				name:    "nombre demasiado largo",
				body:    map[string]string{"nombre": "1234567890123456789012345678901"},
				message: "Key: 'PayloadCategoria.Nombre' Error:Field validation for 'Nombre' failed on the 'max' tag",
			},
			{
				name:    "campo desconocido",
				body:    map[string]any{"nombre": "X", "extra": 1},
				message: `json: unknown field "extra"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"), user.Token, tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("errores de ruta", func(t *testing.T) {
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
				path:    testutil.EmpresaPath(empresa.ID, "/categorias/no-uuid"),
				body:    map[string]string{"nombre": "X"},
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "delete con uuid inválido",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(empresa.ID, "/categorias/no-uuid"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "patch inexistente",
				method:  http.MethodPatch,
				path:    testutil.EmpresaPath(empresa.ID, "/categorias/"+uuid.NewString()),
				body:    map[string]string{"nombre": "X"},
				status:  http.StatusNotFound,
				message: "no encontrado",
			},
			{
				name:    "delete inexistente",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(empresa.ID, "/categorias/"+uuid.NewString()),
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

	t.Run("no puede tocar una categoría de otra empresa", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)
		otherCategory := testutil.CreateCategory(t, otherUser, otherEmpresa.ID)

		patch := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/categorias/"+otherCategory.ID.String()), user.Token, map[string]string{
			"nombre": "Intruso",
		})
		testutil.RequireError(t, patch, http.StatusNotFound, "no encontrado")

		del := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/categorias/"+otherCategory.ID.String()), user.Token, nil)
		testutil.RequireError(t, del, http.StatusNotFound, "no encontrado")
	})

	t.Run("un municipal no puede acceder a las categorías", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"), testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/categorias"))
	requireProtected(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/categorias"))
	requireProtected(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/categorias/"+uuid.NewString()))
}
