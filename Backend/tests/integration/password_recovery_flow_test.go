package integration

import (
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/require"
)

// TestPasswordRecoveryFlow recorre el flujo completo de recuperación de
// contraseña: pedir el código por correo, verificarlo y cambiar la contraseña.
func TestPasswordRecoveryFlow(t *testing.T) {
	user := testutil.NewUser(t)

	const nuevaPassword = "NuevaPassword123"

	var codigo string

	t.Run("pide el código de recuperación y lo manda por correo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password", "", map[string]string{
			"email": user.Email,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)

		codigo = testutil.UltimoCodigoDeRecuperacion()
		require.Len(t, codigo, 6, "el código debe tener 6 dígitos")
		require.Regexp(t, `^[0-9]{6}$`, codigo)
	})

	t.Run("rechaza un email que no está registrado", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password", "", map[string]string{
			"email": testutil.UniqueEmail("recuperacion_nadie"),
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "recurso no encontrado")
	})

	t.Run("rechaza un email inválido", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password", "", map[string]string{
			"email": "no-es-email",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "Key: 'PayloadSolicitarRecuperacion.Email' Error:Field validation for 'Email' failed on the 'email' tag")
	})

	t.Run("rechaza un código incorrecto", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/verificar", "", map[string]string{
			"email":  user.Email,
			"codigo": "000000",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "el código de recuperación no es válido")
	})

	t.Run("exige verificar el código antes de confirmar", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/confirmar", "", map[string]string{
			"email":    user.Email,
			"codigo":   codigo,
			"password": nuevaPassword,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "el código de recuperación todavía no fue verificado")
	})

	t.Run("verifica el código recibido por correo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/verificar", "", map[string]string{
			"email":  user.Email,
			"codigo": codigo,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
	})

	t.Run("confirma la recuperación y cambia la contraseña", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/confirmar", "", map[string]string{
			"email":    user.Email,
			"codigo":   codigo,
			"password": nuevaPassword,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
	})

	t.Run("el usuario puede iniciar sesión con la nueva contraseña", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    user.Email,
			"password": nuevaPassword,
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

	t.Run("el código no se puede reusar", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/verificar", "", map[string]string{
			"email":  user.Email,
			"codigo": codigo,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "el código de recuperación ya fue usado")
	})

	t.Run("pedir un código nuevo invalida el anterior", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password", "", map[string]string{
			"email": user.Email,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)

		nuevo := testutil.UltimoCodigoDeRecuperacion()
		require.NotEmpty(t, nuevo)

		res = testutil.Request(t, http.MethodPost, "/v1/autenticacion/recuperar-password/verificar", "", map[string]string{
			"email":  user.Email,
			"codigo": nuevo,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
	})
}
