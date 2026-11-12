package paginacion

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
		fq, err := NuevaConsultaPaginacionIngresos().Parse(httptest.NewRequest("GET", "/ingresos", nil))

		require.NoError(t, err)
		assert.Equal(t, LimitePorDefecto, fq.Limite)
		assert.Equal(t, 0, fq.Desplazamiento)
		assert.Empty(t, fq.Desde)
		assert.Empty(t, fq.Hasta)
	})

	t.Run("lee limite y desplazamiento de la query string", func(t *testing.T) {
		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?limite=7&desplazamiento=14", nil))

		require.NoError(t, err)
		assert.Equal(t, 7, fq.Limite)
		assert.Equal(t, 14, fq.Desplazamiento)
	})

	t.Run("un parametro vacío no pisa el default", func(t *testing.T) {
		// /ingresos?limite= es el caso de un front que arma la URL con un valor
		// sin definir: no es un error, es "no filter".
		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?limite=&desplazamiento=", nil))

		require.NoError(t, err)
		assert.Equal(t, LimitePorDefecto, fq.Limite)
		assert.Equal(t, 0, fq.Desplazamiento)
	})

	t.Run("ignora query params que no son de la paginación", func(t *testing.T) {
		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?orden=monto&pagina=3", nil))

		require.NoError(t, err)
		assert.Equal(t, LimitePorDefecto, fq.Limite)
		assert.Equal(t, 0, fq.Desplazamiento)
	})

	t.Run("lee el rango de fechas AAAA-MM-DD", func(t *testing.T) {
		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?desde=2026-09-01&hasta=2026-09-30", nil))

		require.NoError(t, err)
		assert.Equal(t, "2026-09-01", fq.Desde)
		assert.Equal(t, "2026-09-30", fq.Hasta)
	})

	t.Run("rechaza limite que no es un entero", func(t *testing.T) {
		_, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?limite=abc", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "limite")
	})

	t.Run("rechaza desplazamiento que no es un entero", func(t *testing.T) {
		_, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?desplazamiento=1.5", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "desplazamiento")
	})

	t.Run("rechaza limite fuera de rango", func(t *testing.T) {
		for _, limite := range []string{"0", "-1", "31", "1000"} {
			t.Run(limite, func(t *testing.T) {
				_, err := NuevaConsultaPaginacionIngresos().Parse(
					httptest.NewRequest("GET", "/ingresos?limite="+limite, nil))

				require.Error(t, err)
			})
		}
	})

	t.Run("rechaza desplazamiento negativo", func(t *testing.T) {
		_, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?desplazamiento=-1", nil))

		require.Error(t, err)
	})

	t.Run("acepta los extremos del rango de limite", func(t *testing.T) {
		for _, limite := range []int{1, LimiteMaximo} {
			fq, err := NuevaConsultaPaginacionIngresos().Parse(
				httptest.NewRequest("GET", fmt.Sprintf("/ingresos?limite=%d", limite), nil))

			require.NoError(t, err)
			assert.Equal(t, limite, fq.Limite)
		}
	})

	t.Run("rechaza fechas que no son AAAA-MM-DD", func(t *testing.T) {
		// incomes.fecha es un DATE. Aceptar "2026-09-01T00:00:00" o
		// "30-09-2026" y descartarlos en silencio haría que el filtro nunca se
		// aplicara.
		for _, fecha := range []string{"30-09-2026", "2026/09/01", "2026-13-01", "ayer"} {
			t.Run(fecha, func(t *testing.T) {
				_, err := NuevaConsultaPaginacionIngresos().Parse(
					httptest.NewRequest("GET", "/ingresos?desde="+fecha, nil))

				require.Error(t, err)
				assert.Contains(t, err.Error(), "desde")
			})
		}
	})

	t.Run("rechaza hasta inválido nombrando el parametro", func(t *testing.T) {
		_, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?hasta=no-es-fecha", nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "hasta")
	})

	t.Run("lee categoria_id como UUID", func(t *testing.T) {
		const catID = "b1c2d3e4-5f60-4a7b-8c9d-0e1f2a3b4c5d"

		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?categoria_id="+catID, nil))

		require.NoError(t, err)
		assert.Equal(t, catID, fq.CategoryID)
	})

	t.Run("sin categoria_id el filtro queda vacio", func(t *testing.T) {
		// Vacío significa "sin filtro": el store lo traduce a ($6 = '' OR ...).
		fq, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?categoria_id=", nil))

		require.NoError(t, err)
		assert.Empty(t, fq.CategoryID)
	})

	t.Run("rechaza categoria_id que no es un UUID", func(t *testing.T) {
		// El filtro es por id, no por nombre. Aceptar "Ventas" como categoría
		// devolvería una lista vacía en vez de un error, y el cliente no
		// distinguiría "no hay ingresos" de "mandaste mal el filtro".
		for _, categoria := range []string{"Ventas", "123", "b1c2d3e4"} {
			t.Run(categoria, func(t *testing.T) {
				_, err := NuevaConsultaPaginacionIngresos().Parse(
					httptest.NewRequest("GET", "/ingresos?categoria_id="+categoria, nil))

				require.Error(t, err)
				assert.Contains(t, err.Error(), "categoria_id")
			})
		}
	})

	t.Run("no devuelve una paginación usable si un parametro es inválido", func(t *testing.T) {
		// Con limite=abc el handler tiene que cortar con 400. Si Parse devolviera
		// la struct con el default, el cliente creería que pidió otra página y
		// el filtro de fechas se perdería en el camino.
		_, err := NuevaConsultaPaginacionIngresos().Parse(
			httptest.NewRequest("GET", "/ingresos?limite=abc&desde=2026-09-01", nil))

		require.Error(t, err)
	})
}
