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

// TestCacheFlow cubre el comportamiento de la suite con el cache activo: el
// router de test se arma con los mismos buscadores cacheados que usa el
// backend real (users.BuscadorUsuarioConCache y empresas.BuscadorEmpresaConCache).
func TestCacheFlow(t *testing.T) {
	user := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, user)
	cat := testutil.CreateCategory(t, user, empresa.ID, model.CategoryTypeGasto)

	t.Run("el middleware resuelve la empresa con el cache", func(t *testing.T) {
		// La primera request carga la empresa de la base y la deja cacheada;
		// la segunda sale del cache. Ambas tienen que dar lo mismo.
		first := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireStatus(t, first, http.StatusOK)

		second := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireStatus(t, second, http.StatusOK)

		require.JSONEq(t, string(first.Body), string(second.Body))
	})

	t.Run("el cache no deja pasar a una empresa ajena", func(t *testing.T) {
		// La clave del cache de empresa incluye el usuario: el intruso no
		// puede reusar la entrada cacheada del dueño.
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("renombrar la empresa invalida la entrada cacheada", func(t *testing.T) {
		// Si el rename no invalidara el cache, el middleware seguiría
		// sirviendo la empresa con el nombre viejo.
		// El nombre va en el path, así que hay que escaparlo.
		nuevo := "Empresa Renombrada " + uuid.NewString()[:8]

		// Se warms el cache de la empresa.
		warm := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireStatus(t, warm, http.StatusOK)

		res := testutil.Request(t, http.MethodPatch, testutil.EmpresaPath(empresa.ID, "/"+url.PathEscape(nuevo)), user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		list := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", user.Token, nil)
		testutil.RequireStatus(t, list, http.StatusOK)

		var encontrada bool
		for _, e := range testutil.DecodeData[[]model.Empresa](t, list) {
			if e.ID == empresa.ID {
				encontrada = true
				require.Equal(t, nuevo, e.NombreEmpresa)
			}
		}
		require.True(t, encontrada, "la empresa tiene que seguir apareciendo en el listado")
	})

	t.Run("borrar la empresa invalida la entrada cacheada", func(t *testing.T) {
		// Si el borrado no invalidara el cache, el middleware seguiría
		// dejando entrar a una empresa que ya no existe.
		otherEmpresa := testutil.CreateEmpresa(t, user)
		catB := testutil.CreateCategory(t, user, otherEmpresa.ID, model.CategoryTypeGasto)
		testutil.CreateExpense(t, user, otherEmpresa.ID, 1000, catB.ID)

		warm := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireStatus(t, warm, http.StatusOK)

		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(otherEmpresa.ID, "/"), user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)

		after := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireError(t, after, http.StatusNotFound, "no encontrado")
	})

	t.Run("el gasto de una empresa sigue siendo accesible con el cache activo", func(t *testing.T) {
		created := testutil.CreateExpense(t, user, empresa.ID, 5000, cat.ID)

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/gastos"), user.Token, nil)
		testutil.RequireStatus(t, list, http.StatusOK)

		var encontrado bool
		for _, expense := range testutil.DecodeData[[]model.Expense](t, list) {
			if expense.ID == created.ID {
				encontrado = true
			}
		}
		require.True(t, encontrado, "el gasto tiene que aparecer en el listado")
	})
}
