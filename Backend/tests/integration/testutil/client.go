package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Result struct {
	Status int
	Body   []byte
	Header http.Header
}

func doRequest(method, path, token string, body any) (Result, error) {
	var reader io.Reader

	switch v := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(v)
	case []byte:
		reader = bytes.NewReader(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return Result{}, fmt.Errorf("serializando payload: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	GetEnv().Router.ServeHTTP(rec, req)

	return Result{
		Status: rec.Code,
		Body:   rec.Body.Bytes(),
		Header: rec.Header(),
	}, nil
}

// Request ejecuta la petición contra el router real, en memoria.
// body puede ser nil, un string/[]byte crudo o cualquier valor que se serialice
// a JSON.
func Request(t *testing.T, method, path, token string, body any) Result {
	t.Helper()

	res, err := doRequest(method, path, token, body)
	require.NoError(t, err)

	return res
}

func DecodeData[T any](t *testing.T, r Result) T {
	t.Helper()

	var envelope struct {
		Datos T `json:"datos"`
	}

	require.NoError(t, json.Unmarshal(r.Body, &envelope), "respuesta no es JSON válido: %s", r.Body)

	return envelope.Datos
}

func DecodeError(t *testing.T, r Result) string {
	t.Helper()

	var envelope struct {
		Error string `json:"error"`
	}

	require.NoError(t, json.Unmarshal(r.Body, &envelope), "respuesta no es JSON válido: %s", r.Body)

	return envelope.Error
}

func RequireStatus(t *testing.T, r Result, status int) {
	t.Helper()

	require.Equal(t, status, r.Status, "cuerpo: %s", r.Body)
}

func RequireError(t *testing.T, r Result, status int, message string) {
	t.Helper()

	require.Equal(t, status, r.Status, "cuerpo: %s", r.Body)
	require.Equal(t, message, DecodeError(t, r), "cuerpo: %s", r.Body)
}
