package integration

import (
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/goals"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGoalsFlow(t *testing.T) {
	user := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, user)

	var goal model.Goal

	t.Run("crea una meta diaria", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/goal"), user.Token, map[string]any{
			"nombre":     testutil.UniqueName("Meta Flow"),
			"periodo":    "dia",
			"monto_meta": 100000,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		goal = testutil.DecodeData[model.Goal](t, res)
		require.NotEqual(t, uuid.Nil, goal.ID)
		require.Equal(t, empresa.ID, goal.EmpresaID)
		require.Equal(t, model.GoalPeriodDia, goal.Periodo)
		require.Equal(t, int64(100000), goal.MontoMeta)
	})

	t.Run("lista las metas de la empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		metas := testutil.DecodeData[[]model.Goal](t, res)

		found := false
		for _, g := range metas {
			if g.ID == goal.ID {
				found = true
			}
		}
		require.True(t, found)
	})

	t.Run("obtiene una meta por id", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+goal.ID.String()), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		got := testutil.DecodeData[model.Goal](t, res)
		require.Equal(t, goal.ID, got.ID)
		require.Equal(t, goal.Nombre, got.Nombre)
		require.Equal(t, goal.MontoMeta, got.MontoMeta)
	})

	t.Run("calcula el estado con los ingresos del período", func(t *testing.T) {
		testutil.CreateIncome(t, user, empresa.ID, 25000, nil)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+goal.ID.String()+"/state"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		state := testutil.DecodeData[goals.GoalState](t, res)
		require.Equal(t, int64(25000), state.MontoActual)
		require.InDelta(t, 25, state.PorcentajeActual, 0.0001)
		require.Equal(t, int64(75000), state.MontoPendiente)
		require.InDelta(t, 75, state.PorcentajeFaltante, 0.0001)
	})

	t.Run("el estado no deja pendiente negativo si se supera la meta", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/goal"), user.Token, map[string]any{
			"nombre":     testutil.UniqueName("Meta Superada"),
			"periodo":    "dia",
			"monto_meta": 10000,
		})
		testutil.RequireStatus(t, res, http.StatusCreated)
		overGoal := testutil.DecodeData[model.Goal](t, res)

		stateRes := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+overGoal.ID.String()+"/state"), user.Token, nil)
		testutil.RequireStatus(t, stateRes, http.StatusOK)
		state := testutil.DecodeData[goals.GoalState](t, stateRes)

		require.Equal(t, int64(25000), state.MontoActual)
		require.InDelta(t, 250, state.PorcentajeActual, 0.0001)
		require.Equal(t, int64(0), state.MontoPendiente)
		require.InDelta(t, 0, state.PorcentajeFaltante, 0.0001)
	})

	t.Run("los períodos semanal y mensual incluyen los ingresos de hoy", func(t *testing.T) {
		for _, periodo := range []string{"semana", "mes"} {
			t.Run(periodo, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/goal"), user.Token, map[string]any{
					"nombre":     testutil.UniqueName("Meta " + periodo),
					"periodo":    periodo,
					"monto_meta": 100000,
				})
				testutil.RequireStatus(t, res, http.StatusCreated)
				created := testutil.DecodeData[model.Goal](t, res)

				stateRes := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+created.ID.String()+"/state"), user.Token, nil)
				testutil.RequireStatus(t, stateRes, http.StatusOK)
				state := testutil.DecodeData[goals.GoalState](t, stateRes)
				require.Equal(t, int64(25000), state.MontoActual)
			})
		}
	})

	t.Run("una empresa sin metas recibe una lista vacía", func(t *testing.T) {
		otherEmpresa := testutil.CreateEmpresa(t, user)
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/goal"), user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Empty(t, testutil.DecodeData[[]model.Goal](t, res))
	})

	t.Run("borra una meta", func(t *testing.T) {
		toDelete := testutil.CreateGoal(t, user, empresa.ID, "dia", 50000)

		res := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/goal/"+toDelete.ID.String()), user.Token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Objetivo eliminado", testutil.DecodeData[string](t, res))

		get := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+toDelete.ID.String()), user.Token, nil)
		testutil.RequireError(t, get, http.StatusNotFound, "not found")

		state := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+toDelete.ID.String()+"/state"), user.Token, nil)
		testutil.RequireError(t, state, http.StatusNotFound, "not found")
	})

	t.Run("creación valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "nombre faltante",
				body:    map[string]any{"periodo": "dia", "monto_meta": 1000},
				message: "Key: 'CreateGoalPayload.Nombre' Error:Field validation for 'Nombre' failed on the 'required' tag",
			},
			{
				name:    "nombre demasiado largo",
				body:    map[string]any{"nombre": "1234567890123456789012345678901", "periodo": "dia", "monto_meta": 1000},
				message: "Key: 'CreateGoalPayload.Nombre' Error:Field validation for 'Nombre' failed on the 'max' tag",
			},
			{
				name:    "período faltante",
				body:    map[string]any{"nombre": "X", "monto_meta": 1000},
				message: "Key: 'CreateGoalPayload.Periodo' Error:Field validation for 'Periodo' failed on the 'required' tag",
			},
			{
				name:    "período inválido",
				body:    map[string]any{"nombre": "X", "periodo": "anio", "monto_meta": 1000},
				message: "Key: 'CreateGoalPayload.Periodo' Error:Field validation for 'Periodo' failed on the 'oneof' tag",
			},
			{
				name:    "monto faltante",
				body:    map[string]any{"nombre": "X", "periodo": "dia"},
				message: "Key: 'CreateGoalPayload.Monto_meta' Error:Field validation for 'Monto_meta' failed on the 'required' tag",
			},
			{
				name:    "monto cero",
				body:    map[string]any{"nombre": "X", "periodo": "dia", "monto_meta": 0},
				message: "Key: 'CreateGoalPayload.Monto_meta' Error:Field validation for 'Monto_meta' failed on the 'required' tag",
			},
			{
				name:    "monto negativo",
				body:    map[string]any{"nombre": "X", "periodo": "dia", "monto_meta": -1},
				message: "Key: 'CreateGoalPayload.Monto_meta' Error:Field validation for 'Monto_meta' failed on the 'gt' tag",
			},
			{
				name:    "campo desconocido",
				body:    map[string]any{"nombre": "X", "periodo": "dia", "monto_meta": 1000, "extra": 1},
				message: `json: unknown field "extra"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/goal"), user.Token, tc.body)
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
				name:    "get con uuid inválido",
				method:  http.MethodGet,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/no-uuid"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "state con uuid inválido",
				method:  http.MethodGet,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/no-uuid/state"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "delete con uuid inválido",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/no-uuid"),
				status:  http.StatusBadRequest,
				message: "invalid UUID length: 7",
			},
			{
				name:    "get inexistente",
				method:  http.MethodGet,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()),
				status:  http.StatusNotFound,
				message: "not found",
			},
			{
				name:    "state inexistente",
				method:  http.MethodGet,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()+"/state"),
				status:  http.StatusNotFound,
				message: "not found",
			},
			{
				name:    "delete inexistente",
				method:  http.MethodDelete,
				path:    testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()),
				status:  http.StatusNotFound,
				message: "not found",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, tc.method, tc.path, user.Token, nil)
				testutil.RequireError(t, res, tc.status, tc.message)
			})
		}
	})

	t.Run("no puede tocar metas de otra empresa", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)
		otherGoal := testutil.CreateGoal(t, otherUser, otherEmpresa.ID, "dia", 1000)

		get := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+otherGoal.ID.String()), user.Token, nil)
		testutil.RequireError(t, get, http.StatusNotFound, "not found")

		state := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+otherGoal.ID.String()+"/state"), user.Token, nil)
		testutil.RequireError(t, state, http.StatusNotFound, "not found")

		del := testutil.Request(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/goal/"+otherGoal.ID.String()), user.Token, nil)
		testutil.RequireError(t, del, http.StatusNotFound, "not found")
	})

	t.Run("no puede acceder a la empresa de otro usuario", func(t *testing.T) {
		otherUser := testutil.NewUser(t)
		otherEmpresa := testutil.CreateEmpresa(t, otherUser)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otherEmpresa.ID, "/goal"), user.Token, nil)
		testutil.RequireError(t, res, http.StatusNotFound, "not found")
	})

	t.Run("un municipal no puede acceder a las metas", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal"), testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, res, http.StatusNotFound, "not found")
	})

	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal"))
	requireProtected(t, http.MethodPost, testutil.EmpresaPath(empresa.ID, "/goal"))
	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()))
	requireProtected(t, http.MethodGet, testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()+"/state"))
	requireProtected(t, http.MethodDelete, testutil.EmpresaPath(empresa.ID, "/goal/"+uuid.NewString()))
}
