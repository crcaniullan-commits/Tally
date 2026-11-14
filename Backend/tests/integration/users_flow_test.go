package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/require"
)

func TestUsersFlow(t *testing.T) {
	user := testutil.NewUser(t)
	const updatedPassword = "NewPassword123"
	const secondPassword = "AnotherPassword123"
	updatedName := "Usuario Renombrado"

	t.Run("actualiza nombre y contraseña", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, map[string]string{
			"nombre":   updatedName,
			"password": updatedPassword,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Usuario actualizado con éxito", testutil.DecodeData[string](t, res))
	})

	t.Run("inicia sesión con la contraseña nueva", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": updatedPassword,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	t.Run("la contraseña anterior ya no sirve", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": user.Password,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	})

	t.Run("actualiza solo la contraseña", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, map[string]string{
			"password": secondPassword,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Usuario actualizado con éxito", testutil.DecodeData[string](t, res))
	})

	t.Run("inicia sesión con la segunda contraseña", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": secondPassword,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	t.Run("actualiza solo el nombre", func(t *testing.T) {
		// El nombre se actualiza sin tocar la contraseña: el usuario del
		// contexto lo carga el middleware con ObtenerPorID, que sí trae el
		// hash, así que el UPDATE no escribe NULL en una columna NOT NULL.
		res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, map[string]string{
			"nombre": "Solo Nombre",
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Usuario actualizado con éxito", testutil.DecodeData[string](t, res))

		var nombre string
		require.NoError(t, testutil.GetEnv().DB.QueryRow(`SELECT nombre FROM users WHERE email = $1`, user.Email).Scan(&nombre))
		require.Equal(t, "Solo Nombre", nombre)
	})

	t.Run("un payload vacío no rompe nada", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, map[string]any{})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Usuario actualizado con éxito", testutil.DecodeData[string](t, res))
	})

	t.Run("actualizar solo el nombre deja la contraseña intacta", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, map[string]string{
			"nombre": "Nombre Con Password",
		})

		testutil.RequireStatus(t, res, http.StatusOK)

		login := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": secondPassword,
		})

		testutil.RequireStatus(t, login, http.StatusAccepted)
	})

	t.Run("un municipal consulta un usuario por RUT", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/usuarios/municipal/"+user.RUT, testutil.MunicipalToken(t), nil)

		testutil.RequireStatus(t, res, http.StatusOK)

		data := testutil.DecodeData[map[string]any](t, res)
		require.Equal(t, user.ID.String(), data["id"])
		require.Equal(t, user.Email, data["email"])

		// El nombre es el que tenga el usuario en la base: otros subtests del
		// flujo lo cambian.
		var nombreActual string
		require.NoError(t, testutil.GetEnv().DB.QueryRow(`SELECT nombre FROM users WHERE email = $1`, user.Email).Scan(&nombreActual))
		require.Equal(t, nombreActual, data["nombre"])

		require.Equal(t, "usuario", data["role"])
		require.NotNil(t, data["plan_expires_at"])

		rut, ok := data["rut"].(map[string]any)
		require.True(t, ok, "rut debería ser un objeto: %v", data["rut"])
		// DISCREPANCIA: el swagger define rut como {cuerpo, dv}, pero el struct
		// model.RUT no tiene tags JSON y serializa {Cuerpo, DV}.
		require.Contains(t, rut, "Cuerpo")
		require.Contains(t, rut, "DV")
		require.NotContains(t, rut, "cuerpo")
		require.NotContains(t, rut, "dv")

		// BUG: ObtenerPorRUT no selecciona created_at ni updated_at, así que salen
		// en el zero value de time.Time.
		require.Equal(t, "0001-01-01T00:00:00Z", data["created_at"])
		require.Equal(t, "0001-01-01T00:00:00Z", data["updated_at"])
	})

	t.Run("consulta por RUT con formato inválido", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/usuarios/municipal/abc", testutil.MunicipalToken(t), nil)

		testutil.RequireError(t, res, http.StatusBadRequest, "el formato del RUT no es válido")
	})

	t.Run("consulta por RUT inexistente", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/usuarios/municipal/99999999-9", testutil.MunicipalToken(t), nil)

		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("un usuario común no puede consultar por RUT", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/usuarios/municipal/"+user.RUT, user.Token, nil)

		testutil.RequireError(t, res, http.StatusForbidden, "prohibido")
	})

	t.Run("un municipal no puede actualizar ni borrar usuarios", func(t *testing.T) {
		patch := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", testutil.MunicipalToken(t), map[string]string{
			"nombre": "Municipal Intruso",
		})
		testutil.RequireError(t, patch, http.StatusForbidden, "prohibido")

		del := testutil.Request(t, http.MethodDelete, "/v1/aplicacion/usuarios", testutil.MunicipalToken(t), nil)
		testutil.RequireError(t, del, http.StatusForbidden, "prohibido")
	})

	t.Run("validación del payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "contraseña corta",
				body:    map[string]string{"password": "short"},
				message: "Key: 'PayloadActualizarUsuario.Password' Error:Field validation for 'Password' failed on the 'min' tag",
			},
			{
				name:    "contraseña demasiado larga",
				body:    map[string]string{"password": strings.Repeat("a", 73)},
				message: "Key: 'PayloadActualizarUsuario.Password' Error:Field validation for 'Password' failed on the 'max' tag",
			},
			{
				name:    "nombre demasiado largo",
				body:    map[string]string{"nombre": strings.Repeat("n", 101)},
				message: "Key: 'PayloadActualizarUsuario.Nombre' Error:Field validation for 'Nombre' failed on the 'max' tag",
			},
			{
				name:    "campo desconocido",
				body:    map[string]any{"nombre": "X", "extra": 1},
				message: `json: unknown field "extra"`,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPatch, "/v1/aplicacion/usuarios", user.Token, tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("elimina su propia cuenta", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, "/v1/aplicacion/usuarios", user.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Usuario eliminado", testutil.DecodeData[string](t, res))
	})

	t.Run("el token deja de servir tras borrar la cuenta", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", user.Token, nil)

		testutil.RequireError(t, res, http.StatusUnauthorized, "no autorizado")
	})

	t.Run("no se puede volver a iniciar sesión", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": secondPassword,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "recurso no encontrado")
	})

	requireProtected(t, http.MethodPatch, "/v1/aplicacion/usuarios")
	requireProtected(t, http.MethodDelete, "/v1/aplicacion/usuarios")
	requireProtected(t, http.MethodGet, "/v1/aplicacion/usuarios/municipal/99999999-9")
}
