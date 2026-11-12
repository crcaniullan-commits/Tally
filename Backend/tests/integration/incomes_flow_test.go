package integration

import (
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIncomesFlow(t *testing.T) {
	user := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, user)
	catA := testutil.CreateCategory(t, user, empresa.ID)
	catB := testutil.CreateCategory(t, user, empresa.ID)

	var conCategoria, sinCategoria, conGlobal model.Income

	t.Run("crea un ingreso con categoría y descripción", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          15000,
			"payment_method": "debito",
			"descripcion":    "venta de prueba",
			"category_id":    catA.ID.String(),
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		conCategoria = testutil.DecodeData[model.Income](t, res)
		require.NotEqual(t, uuid.Nil, conCategoria.ID)
		require.Equal(t, empresa.ID, conCategoria.EmpresaID)
		require.Equal(t, int64(15000), conCategoria.Monto)
		require.Equal(t, model.PaymentMethodDebito, conCategoria.PaymentMethod)
		require.NotNil(t, conCategoria.Descripcion)
		require.Equal(t, "venta de prueba", *conCategoria.Descripcion)
		require.NotNil(t, conCategoria.CategoryID)
		require.Equal(t, catA.ID, *conCategoria.CategoryID)
		require.False(t, conCategoria.Fecha.IsZero())
	})

	t.Run("crea un ingreso sin categoría", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          5000,
			"payment_method": "efectivo",
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		sinCategoria = testutil.DecodeData[model.Income](t, res)
		require.Nil(t, sinCategoria.CategoryID)
	})

	t.Run("crea un ingreso con una categoría global", func(t *testing.T) {
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

		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          7000,
			"payment_method": "transferencia",
			"category_id":    global.ID.String(),
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		conGlobal = testutil.DecodeData[model.Income](t, res)
		require.NotNil(t, conGlobal.CategoryID)
		require.Equal(t, global.ID, *conGlobal.CategoryID)
	})

	t.Run("lista los ingresos de la empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		incomes := testutil.DecodeData[[]model.Income](t, res)

		ids := map[uuid.UUID]bool{}
		for _, income := range incomes {
			ids[income.ID] = true
		}
		require.True(t, ids[conCategoria.ID])
		require.True(t, ids[sinCategoria.ID])
		require.True(t, ids[conGlobal.ID])
	})

	t.Run("filtra por categoría", func(t *testing.T) {
		otherEmpresa := testutil.CreateEmpresa(t, user)
		catX := testutil.CreateCategory(t, user, otherEmpresa.ID)
		catY := testutil.CreateCategory(t, user, otherEmpresa.ID)

		withX := testutil.CreateIncome(t, user, otherEmpresa.ID, 1000, &catX.ID)
		testutil.CreateIncome(t, user, otherEmpresa.ID, 2000, &catY.ID)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?categoria_id="+catX.ID.String()), user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		incomes := testutil.DecodeData[[]model.Income](t, res)

		require.Len(t, incomes, 1)
		require.Equal(t, withX.ID, incomes[0].ID)
	})

	t.Run("filtra por rango de fechas", func(t *testing.T) {
		otherEmpresa := testutil.CreateEmpresa(t, user)
		created := testutil.CreateIncome(t, user, otherEmpresa.ID, 3000, nil)

		todayRes := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?desde="+dbDate(t, 0)+"&hasta="+dbDate(t, 0)), user.Token, nil)
		testutil.RequireStatus(t, todayRes, http.StatusOK)
		todayIncomes := testutil.DecodeData[[]model.Income](t, todayRes)
		require.Len(t, todayIncomes, 1)
		require.Equal(t, created.ID, todayIncomes[0].ID)

		futureRes := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?desde="+dbDate(t, 1)), user.Token, nil)
		testutil.RequireStatus(t, futureRes, http.StatusOK)
		require.Empty(t, testutil.DecodeData[[]model.Income](t, futureRes))

		pastRes := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?hasta="+dbDate(t, -1)), user.Token, nil)
		testutil.RequireStatus(t, pastRes, http.StatusOK)
		require.Empty(t, testutil.DecodeData[[]model.Income](t, pastRes))
	})

	t.Run("pagina los resultados", func(t *testing.T) {
		otherEmpresa := testutil.CreateEmpresa(t, user)
		for i := 0; i < 3; i++ {
			testutil.CreateIncome(t, user, otherEmpresa.ID, int64(1000*(i+1)), nil)
		}

		first := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?limite=2"), user.Token, nil)
		testutil.RequireStatus(t, first, http.StatusOK)
		require.Len(t, testutil.DecodeData[[]model.Income](t, first), 2)

		second := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?limite=2&desplazamiento=2"), user.Token, nil)
		testutil.RequireStatus(t, second, http.StatusOK)
		require.Len(t, testutil.DecodeData[[]model.Income](t, second), 1)

		all := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos"), user.Token, nil)
		testutil.RequireStatus(t, all, http.StatusOK)
		require.Len(t, testutil.DecodeData[[]model.Income](t, all), 3)

		beyond := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos?desplazamiento=10"), user.Token, nil)
		testutil.RequireStatus(t, beyond, http.StatusOK)
		require.Empty(t, testutil.DecodeData[[]model.Income](t, beyond))
	})

	t.Run("borra un ingreso", func(t *testing.T) {
		toDelete := testutil.CreateIncome(t, user, empresa.ID, 1234, &catB.ID)

		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/ingresos/"+toDelete.ID.String()), user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "ingreso eliminado", testutil.DecodeData[string](t, res))

		list := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, nil)
		incomes := testutil.DecodeData[[]model.Income](t, list)
		for _, income := range incomes {
			require.NotEqual(t, toDelete.ID, income.ID)
		}

		again := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/ingresos/"+toDelete.ID.String()), user.Token, nil)
		testutil.RequireError(t, again, http.StatusNotFound, "no encontrado")
	})

	t.Run("creación valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "monto faltante",
				body:    map[string]any{"payment_method": "efectivo"},
				message: "Key: 'PayloadIngreso.Monto' Error:Field validation for 'Monto' failed on the 'required' tag",
			},
			{
				name:    "monto cero",
				body:    map[string]any{"monto": 0, "payment_method": "efectivo"},
				message: "Key: 'PayloadIngreso.Monto' Error:Field validation for 'Monto' failed on the 'required' tag",
			},
			{
				name:    "monto negativo",
				body:    map[string]any{"monto": -1, "payment_method": "efectivo"},
				message: "Key: 'PayloadIngreso.Monto' Error:Field validation for 'Monto' failed on the 'gt' tag",
			},
			{
				name:    "monto string",
				body:    `{"monto":"1000","payment_method":"efectivo"}`,
				message: "json: cannot unmarshal string into Go struct field PayloadIngreso.monto of type int64",
			},
			{
				name:    "monto decimal",
				body:    `{"monto":1000.5,"payment_method":"efectivo"}`,
				message: "json: cannot unmarshal number 1000.5 into Go struct field PayloadIngreso.monto of type int64",
			},
			{
				name:    "payment method inválido",
				body:    map[string]any{"monto": 1000, "payment_method": "bitcoin"},
				message: "Key: 'PayloadIngreso.PaymentMethod' Error:Field validation for 'PaymentMethod' failed on the 'oneof' tag",
			},
			{
				name:    "payment method faltante",
				body:    map[string]any{"monto": 1000},
				message: "Key: 'PayloadIngreso.PaymentMethod' Error:Field validation for 'PaymentMethod' failed on the 'required' tag",
			},
			{
				name:    "descripción demasiado larga",
				body:    map[string]any{"monto": 1000, "payment_method": "efectivo", "descripcion": string(make([]byte, 101))},
				message: "Key: 'PayloadIngreso.Descripcion' Error:Field validation for 'Descripcion' failed on the 'max' tag",
			},
			{
				name:    "campo desconocido",
				body:    map[string]any{"monto": 1000, "payment_method": "efectivo", "fecha": "2026-01-01"},
				message: `json: unknown field "fecha"`,
			},
			{
				name:    "categoria_id inválido",
				body:    map[string]any{"monto": 1000, "payment_method": "efectivo", "category_id": "nope"},
				message: "invalid UUID length: 4",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("rechaza una categoría inexistente", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          1000,
			"payment_method": "efectivo",
			"category_id":    uuid.NewString(),
		})
		testutil.RequireError(t, res, http.StatusBadRequest, "recurso no encontrado")
	})

	t.Run("rechaza una categoría de otra empresa", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)
		otherCategory := testutil.CreateCategory(t, otherUser, otherEmpresa.ID)

		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          1000,
			"payment_method": "efectivo",
			"category_id":    otherCategory.ID.String(),
		})
		testutil.RequireError(t, res, http.StatusBadRequest, "recurso no encontrado")
	})

	t.Run("valida los query params de paginación", func(t *testing.T) {
		cases := []struct {
			name    string
			query   string
			message string
		}{
			{
				name:    "limit 0",
				query:   "limite=0",
				message: "Key: 'ConsultaPaginacionIngresos.Limite' Error:Field validation for 'Limite' failed on the 'gte' tag",
			},
			{
				name:    "limit 31",
				query:   "limite=31",
				message: "Key: 'ConsultaPaginacionIngresos.Limite' Error:Field validation for 'Limite' failed on the 'lte' tag",
			},
			{
				name:    "limit no entero",
				query:   "limite=abc",
				message: `limite inválido: "abc" no es un entero`,
			},
			{
				name:    "offset negativo",
				query:   "desplazamiento=-1",
				message: "Key: 'ConsultaPaginacionIngresos.Desplazamiento' Error:Field validation for 'Desplazamiento' failed on the 'gte' tag",
			},
			{
				name:    "offset no entero",
				query:   "desplazamiento=abc",
				message: `desplazamiento inválido: "abc" no es un entero`,
			},
			{
				name:    "desde inválido",
				query:   "desde=2026-13-01",
				message: `desde inválido: "2026-13-01" no es una fecha AAAA-MM-DD`,
			},
			{
				name:    "hasta inválido",
				query:   "hasta=01-01-2026",
				message: `hasta inválido: "01-01-2026" no es una fecha AAAA-MM-DD`,
			},
			{
				name:    "categoria_id inválido",
				query:   "categoria_id=nope",
				message: `categoria_id inválido: "nope" no es un UUID`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/ingresos?"+tc.query), user.Token, nil)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
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
				path:    testutil.EmpresaPath(empresa.ID, "/ingresos/no-uuid"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "delete inexistente",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(empresa.ID, "/ingresos/"+uuid.NewString()),
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

	t.Run("no puede acceder a los ingresos de otra empresa", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)

		get := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos"), user.Token, nil)
		testutil.RequireError(t, get, http.StatusNotFound, "no encontrado")

		post := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(otherEmpresa.ID, "/ingresos"), user.Token, map[string]any{
			"monto":          1000,
			"payment_method": "efectivo",
		})
		testutil.RequireError(t, post, http.StatusNotFound, "no encontrado")
	})

	t.Run("un municipal no puede acceder a los ingresos", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/ingresos"), testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/ingresos"))
	requireProtected(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/ingresos"))
	requireProtected(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/ingresos/"+uuid.NewString()))
}
