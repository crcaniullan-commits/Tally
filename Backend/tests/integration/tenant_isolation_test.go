package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/require"
)

// TestTenantIsolation verifica que la tenencia por empresa se aplique en el
// WHERE de cada store y no solo en el middleware: dos usuarios con empresas
// distintas no pueden leer ni modificar datos ajenos, ni siquiera conociendo
// los IDs.
func TestTenantIsolation(t *testing.T) {
	owner := testutil.NewUser(t)
	ownerEmpresa := testutil.CreateEmpresa(t, owner)
	ownerCat := testutil.CreateCategory(t, owner, ownerEmpresa.ID)
	ownerIncome := testutil.CreateIncome(t, owner, ownerEmpresa.ID, 1500, &ownerCat.ID)
	ownerExpense := testutil.CreateExpense(t, owner, ownerEmpresa.ID, 700, ownerCat.ID)
	ownerDebtor := testutil.CreateDebtor(t, owner, ownerEmpresa.ID, 2000)
	ownerGoal := testutil.CreateGoal(t, owner, ownerEmpresa.ID, "dia", 10000)

	intruder := testutil.NewUser(t)
	intruderEmpresa := testutil.CreateEmpresa(t, intruder)
	intruderCat := testutil.CreateCategory(t, intruder, intruderEmpresa.ID)
	intruderIncome := testutil.CreateIncome(t, intruder, intruderEmpresa.ID, 900, &intruderCat.ID)
	intruderExpense := testutil.CreateExpense(t, intruder, intruderEmpresa.ID, 400, intruderCat.ID)
	intruderDebtor := testutil.CreateDebtor(t, intruder, intruderEmpresa.ID, 300)
	intruderGoal := testutil.CreateGoal(t, intruder, intruderEmpresa.ID, "dia", 5000)

	t.Run("el intruso no puede entrar a la empresa ajena", func(t *testing.T) {
		cases := []struct {
			name   string
			method string
			path   string
			body   any
		}{
			{name: "listar ingresos", method: http.MethodGet, path: testutil.EmpresaPath(ownerEmpresa.ID, "/ingresos")},
			{name: "listar gastos", method: http.MethodGet, path: testutil.EmpresaPath(ownerEmpresa.ID, "/gastos")},
			{name: "listar categorías", method: http.MethodGet, path: testutil.EmpresaPath(ownerEmpresa.ID, "/categorias")},
			{name: "listar metas", method: http.MethodGet, path: testutil.EmpresaPath(ownerEmpresa.ID, "/metas")},
			{name: "listar deudores", method: http.MethodGet, path: testutil.EmpresaPath(ownerEmpresa.ID, "/deudas")},
			{
				name:   "renombrar la empresa",
				method: http.MethodPatch,
				path:   testutil.EmpresaPath(ownerEmpresa.ID, "/Intruso"),
			},
			{name: "borrar la empresa", method: http.MethodDelete, path: testutil.EmpresaPath(ownerEmpresa.ID, "")},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, tc.method, tc.path, intruder.Token, tc.body)
				testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
			})
		}
	})

	t.Run("no puede borrar recursos ajenos desde su propia empresa", func(t *testing.T) {
		cases := []struct {
			name   string
			method string
			path   string
			body   any
		}{
			{
				name:   "ingreso ajeno",
				method: http.MethodDelete,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/ingresos/"+ownerIncome.ID.String()),
			},
			{
				name:   "gasto ajeno",
				method: http.MethodDelete,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/gastos/"+ownerExpense.ID.String()),
			},
			{
				name:   "meta ajena",
				method: http.MethodGet,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/metas/"+ownerGoal.ID.String()),
			},
			{
				name:   "estado de meta ajena",
				method: http.MethodGet,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/metas/"+ownerGoal.ID.String()+"/estado"),
			},
			{
				name:   "borrar meta ajena",
				method: http.MethodDelete,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/metas/"+ownerGoal.ID.String()),
			},
			{
				name:   "actualizar deudor ajeno",
				method: http.MethodPatch,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/deudas/"+ownerDebtor.ID.String()),
				body: map[string]any{
					"nombre":       "Intruso",
					"monto":        1000,
					"fecha_limite": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
				},
			},
			{
				name:   "marcar deudor ajeno como pagado",
				method: http.MethodPatch,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/deudas/"+ownerDebtor.ID.String()+"/pagada"),
			},
			{
				name:   "renombrar categoría ajena",
				method: http.MethodPatch,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/categorias/"+ownerCat.ID.String()),
				body:   map[string]string{"nombre": testutil.UniqueName("Intrusa")},
			},
			{
				name:   "borrar categoría ajena",
				method: http.MethodDelete,
				path:   testutil.EmpresaPath(intruderEmpresa.ID, "/categorias/"+ownerCat.ID.String()),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, tc.method, tc.path, intruder.Token, tc.body)
				testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
			})
		}
	})

	t.Run("no puede usar una categoría de otra empresa", func(t *testing.T) {
		income := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(intruderEmpresa.ID, "/ingresos"), intruder.Token, map[string]any{
			"monto":          1000,
			"payment_method": "efectivo",
			"category_id":    ownerCat.ID.String(),
		})
		testutil.RequireError(t, income, http.StatusBadRequest, "recurso no encontrado")

		expense := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(intruderEmpresa.ID, "/gastos"), intruder.Token, map[string]any{
			"monto":       1000,
			"category_id": ownerCat.ID.String(),
		})
		testutil.RequireError(t, expense, http.StatusBadRequest, "recurso no encontrado")
	})

	t.Run("los listados del intruso solo traen sus datos", func(t *testing.T) {
		incomes := testutil.DecodeData[[]model.Income](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(intruderEmpresa.ID, "/ingresos"), intruder.Token, nil))
		require.Len(t, incomes, 1)
		require.Equal(t, intruderIncome.ID, incomes[0].ID)

		expenses := testutil.DecodeData[[]model.Expense](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(intruderEmpresa.ID, "/gastos"), intruder.Token, nil))
		require.Len(t, expenses, 1)
		require.Equal(t, intruderExpense.ID, expenses[0].ID)

		debtors := testutil.DecodeData[[]model.Debtor](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(intruderEmpresa.ID, "/deudas"), intruder.Token, nil))
		require.Len(t, debtors, 1)
		require.Equal(t, intruderDebtor.ID, debtors[0].ID)

		goals := testutil.DecodeData[[]model.Goal](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(intruderEmpresa.ID, "/metas"), intruder.Token, nil))
		require.Len(t, goals, 1)
		require.Equal(t, intruderGoal.ID, goals[0].ID)

		categorias := testutil.DecodeData[[]model.Categorie](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(intruderEmpresa.ID, "/categorias"), intruder.Token, nil))
		for _, c := range categorias {
			require.NotEqual(t, ownerCat.ID, c.ID)
		}
		require.Contains(t, categorias, intruderCat)
	})

	t.Run("el dueño sigue viendo sus propios datos", func(t *testing.T) {
		incomes := testutil.DecodeData[[]model.Income](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/ingresos"), owner.Token, nil))
		require.Len(t, incomes, 1)
		require.Equal(t, ownerIncome.ID, incomes[0].ID)

		expenses := testutil.DecodeData[[]model.Expense](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/gastos"), owner.Token, nil))
		require.Len(t, expenses, 1)
		require.Equal(t, ownerExpense.ID, expenses[0].ID)

		debtors := testutil.DecodeData[[]model.Debtor](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/deudas"), owner.Token, nil))
		require.Len(t, debtors, 1)
		require.Equal(t, ownerDebtor.ID, debtors[0].ID)

		goals := testutil.DecodeData[[]model.Goal](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/metas"), owner.Token, nil))
		require.Len(t, goals, 1)
		require.Equal(t, ownerGoal.ID, goals[0].ID)

		categorias := testutil.DecodeData[[]model.Categorie](t, testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/categorias"), owner.Token, nil))
		require.Contains(t, categorias, ownerCat)
		for _, c := range categorias {
			require.NotEqual(t, intruderCat.ID, c.ID)
		}
	})

	t.Run("un municipal no accede a las empresas de los usuarios", func(t *testing.T) {
		// EmpresaOfUser busca por (empresaID, userID): el municipal no tiene
		// empresas, así que el middleware responde 404 antes del chequeo de rol.
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(ownerEmpresa.ID, "/ingresos"), testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")

		list := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, list, http.StatusForbidden, "prohibido")
	})

}
