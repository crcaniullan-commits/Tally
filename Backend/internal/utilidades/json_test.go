package utilidades

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readJSONAndValidatePayload struct {
	Nombre string `json:"nombre" validate:"required,max=10"`
	Edad   int    `json:"edad" validate:"required,gt=0"`
}

func TestReadJSONAndValidate(t *testing.T) {
	t.Run("payload valido", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"nombre":"Tally","edad":3}`))
		w := httptest.NewRecorder()

		var payload readJSONAndValidatePayload

		require.NoError(t, LeerJSONYValidar(w, r, &payload))
		assert.Equal(t, "Tally", payload.Nombre)
		assert.Equal(t, 3, payload.Edad)
	})

	t.Run("json invalido", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"nombre":`))
		w := httptest.NewRecorder()

		var payload readJSONAndValidatePayload

		assert.Error(t, LeerJSONYValidar(w, r, &payload))
	})

	t.Run("campo desconocido", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"nombre":"Tally","edad":3,"extra":1}`))
		w := httptest.NewRecorder()

		var payload readJSONAndValidatePayload

		assert.Error(t, LeerJSONYValidar(w, r, &payload))
	})

	t.Run("validacion fallida", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"nombre":"","edad":0}`))
		w := httptest.NewRecorder()

		var payload readJSONAndValidatePayload

		assert.Error(t, LeerJSONYValidar(w, r, &payload))
	})
}
