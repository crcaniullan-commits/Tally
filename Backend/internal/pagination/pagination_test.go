package pagination

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Los tests arman la request con httptest.NewRequest: a Parse solo le interesa
// r.URL.Query().

func TestIncomePaginationQuery_Parse(t *testing.T) {
	t.Run("sin query string devuelve la primera página del tamaño tope", func(t *testing.T) {
		fq, err := NewIncomePaginationQuery().Parse(httptest.NewRequest("GET", "/incomes", nil))

		require.NoError(t, err)
		assert.Equal(t, DefaultLimit, fq.Limit)
		assert.Equal(t, 0, fq.Offset)
		assert.Empty(t, fq.Since)
		assert.Empty(t, fq.Until)
	})

	t.Run("lee limit y offset de la query string", func(t *testing.T) {
		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?limit=7&offset=14", nil))

		require.NoError(t, err)
		assert.Equal(t, 7, fq.Limit)
		assert.Equal(t, 14, fq.Offset)
	})

	t.Run("un parametro vacío no pisa el default", func(t *testing.T) {
		// /incomes?limit= es el caso de un front que arma la URL con un valor
		// sin definir: no es un error, es "no filter".
		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?limit=&offset=", nil))

		require.NoError(t, err)
		assert.Equal(t, DefaultLimit, fq.Limit)
		assert.Equal(t, 0, fq.Offset)
	})

	t.Run("ignora query params que no son de la paginación", func(t *testing.T) {
		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?orden=monto&page=3", nil))

		require.NoError(t, err)
		assert.Equal(t, DefaultLimit, fq.Limit)
		assert.Equal(t, 0, fq.Offset)
	})

	t.Run("lee el rango de fechas AAAA-MM-DD", func(t *testing.T) {
		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?since=2026-09-01&until=2026-09-30", nil))

		require.NoError(t, err)
		assert.Equal(t, "2026-09-01", fq.Since)
		assert.Equal(t, "2026-09-30", fq.Until)
	})

	t.Run("rechaza limit que no es un entero", func(t *testing.T) {
		_, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?limit=abc", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "limit")
	})

	t.Run("rechaza offset que no es un entero", func(t *testing.T) {
		_, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?offset=1.5", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "offset")
	})

	t.Run("rechaza limit fuera de rango", func(t *testing.T) {
		for _, limit := range []string{"0", "-1", "31", "1000"} {
			t.Run(limit, func(t *testing.T) {
				_, err := NewIncomePaginationQuery().Parse(
					httptest.NewRequest("GET", "/incomes?limit="+limit, nil))

				require.Error(t, err)
			})
		}
	})

	t.Run("rechaza offset negativo", func(t *testing.T) {
		_, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?offset=-1", nil))

		require.Error(t, err)
	})

	t.Run("acepta los extremos del rango de limit", func(t *testing.T) {
		for _, limite := range []int{1, MaxLimit} {
			fq, err := NewIncomePaginationQuery().Parse(
				httptest.NewRequest("GET", fmt.Sprintf("/incomes?limit=%d", limite), nil))

			require.NoError(t, err)
			assert.Equal(t, limite, fq.Limit)
		}
	})

	t.Run("rechaza fechas que no son AAAA-MM-DD", func(t *testing.T) {
		// incomes.fecha es un DATE. Aceptar "2026-09-01T00:00:00" o
		// "30-09-2026" y descartarlos en silencio haría que el filtro nunca se
		// aplicara.
		for _, fecha := range []string{"30-09-2026", "2026/09/01", "2026-13-01", "ayer"} {
			t.Run(fecha, func(t *testing.T) {
				_, err := NewIncomePaginationQuery().Parse(
					httptest.NewRequest("GET", "/incomes?since="+fecha, nil))

				require.Error(t, err)
				assert.Contains(t, err.Error(), "since")
			})
		}
	})

	t.Run("rechaza until inválido nombrando el parametro", func(t *testing.T) {
		_, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?until=no-es-fecha", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "until")
	})

	t.Run("lee category_id como UUID", func(t *testing.T) {
		const catID = "b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"

		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?category_id="+catID, nil))

		require.NoError(t, err)
		assert.Equal(t, catID, fq.CategoryID)
	})

	t.Run("sin category_id el filtro queda vacio", func(t *testing.T) {
		// Vacío significa "sin filtro": el store lo traduce a ($6 = '' OR ...).
		fq, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?category_id=", nil))

		require.NoError(t, err)
		assert.Empty(t, fq.CategoryID)
	})

	t.Run("rechaza category_id que no es un UUID", func(t *testing.T) {
		// El filtro es por id, no por nombre. Aceptar "Ventas" como categoría
		// devolvería una lista vacía en vez de un error, y el cliente no
		// distinguiría "no hay ingresos" de "mandaste mal el filtro".
		for _, categoria := range []string{"Ventas", "123", "b1c2d3e4"} {
			t.Run(categoria, func(t *testing.T) {
				_, err := NewIncomePaginationQuery().Parse(
					httptest.NewRequest("GET", "/incomes?category_id="+categoria, nil))

				require.Error(t, err)
				assert.Contains(t, err.Error(), "category_id")
			})
		}
	})

	t.Run("no devuelve una paginación usable si un parametro es inválido", func(t *testing.T) {
		// Con limit=abc el handler tiene que cortar con 400. Si Parse devolviera
		// la struct con el default, el cliente creería que pidió otra página y
		// el filtro de fechas se perdería en el camino.
		_, err := NewIncomePaginationQuery().Parse(
			httptest.NewRequest("GET", "/incomes?limit=abc&since=2026-09-01", nil))

		require.Error(t, err)
	})
}
