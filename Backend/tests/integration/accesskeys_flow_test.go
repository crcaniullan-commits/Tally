package integration

import (
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccessKeysFlow(t *testing.T) {
	municipal := testutil.MunicipalToken(t)
	email := testutil.UniqueEmail("key_target")

	var key model.AccessKey

	t.Run("el municipal emite una llave", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key", municipal, map[string]string{
			"email": email,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		key = testutil.DecodeData[model.AccessKey](t, res)
		require.NotEqual(t, uuid.Nil, key.ID)
		require.Len(t, key.Code, 6)
		require.NotNil(t, key.CreatedBy)
		require.NotNil(t, key.ExpiresAt)
	})

	t.Run("el municipal reenvía la llave por id", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/AccessID/"+key.ID.String(), municipal, map[string]string{
			"email": email,
		})

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Correo reenviado", testutil.DecodeData[string](t, res))
	})

	t.Run("el municipal revoca la llave por código", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/revoke/"+key.Code, municipal, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		require.Equal(t, "Codigo desabilitado", testutil.DecodeData[string](t, res))
	})

	t.Run("una llave revocada ya no se puede canjear", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/auth", "", map[string]string{
			"email":    testutil.UniqueEmail("key_revoked"),
			"password": "Password123",
			"nombre":   "Llave Revocada",
			"rut":      testutil.NewRUT(),
			"code":     key.Code,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "The code can't be Redeemed")
	})

	t.Run("una llave nueva sí se puede canjear", func(t *testing.T) {
		freshEmail := testutil.UniqueEmail("key_fresh")
		fresh := testutil.IssueKey(t, freshEmail)
		res := testutil.Request(t, http.MethodPost, "/v1/auth", "", map[string]string{
			"email":    freshEmail,
			"password": "Password123",
			"nombre":   "Llave Nueva",
			"rut":      testutil.NewRUT(),
			"code":     fresh.Code,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	t.Run("un usuario normal no puede emitir llaves", func(t *testing.T) {
		u := testutil.NewUser(t)
		res := testutil.Request(t, http.MethodPost, "/v1/app/key", u.Token, map[string]string{
			"email": testutil.UniqueEmail("key_forbidden"),
		})

		testutil.RequireError(t, res, http.StatusForbidden, "forbidden")
	})

	t.Run("un usuario normal no puede reenviar llaves", func(t *testing.T) {
		u := testutil.NewUser(t)
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/AccessID/"+uuid.NewString(), u.Token, map[string]string{
			"email": testutil.UniqueEmail("key_forbidden"),
		})

		testutil.RequireError(t, res, http.StatusForbidden, "forbidden")
	})

	t.Run("un usuario normal no puede revocar llaves", func(t *testing.T) {
		u := testutil.NewUser(t)
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/revoke/cualquier-codigo", u.Token, nil)

		testutil.RequireError(t, res, http.StatusForbidden, "forbidden")
	})

	t.Run("valida el email al emitir", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key", municipal, map[string]string{
			"email": "no-es-email",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "Key: 'EmailPayload.Email' Error:Field validation for 'Email' failed on the 'email' tag")
	})

	t.Run("valida el email al reenviar", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/AccessID/"+key.ID.String(), municipal, map[string]string{
			"email": "no-es-email",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "Key: 'EmailPayload.Email' Error:Field validation for 'Email' failed on the 'email' tag")
	})

	t.Run("reenviar con un id inválido devuelve 400", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/AccessID/not-a-uuid", municipal, map[string]string{
			"email": email,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "invalid UUID length: 10")
	})

	t.Run("reenviar una llave inexistente devuelve 404", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/AccessID/"+uuid.NewString(), municipal, map[string]string{
			"email": email,
		})

		testutil.RequireError(t, res, http.StatusNotFound, "not found")
	})

	t.Run("revocar un código inexistente devuelve 404", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/app/key/revoke/no-existe", municipal, nil)

		testutil.RequireError(t, res, http.StatusNotFound, "not found")
	})

	requireProtected(t, http.MethodPost, "/v1/app/key")
	requireProtected(t, http.MethodPost, "/v1/app/key/AccessID/"+uuid.NewString())
	requireProtected(t, http.MethodPost, "/v1/app/key/revoke/cualquier-codigo")
}
