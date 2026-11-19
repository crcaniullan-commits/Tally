package cache

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cacheFalso es un Cache en memoria para probar el package sin Redis.
// Guarda lo mismo que guardaría el real: los bytes tal cual, sin interpretar.
type cacheFalso struct {
	values map[string][]byte
	// getCalls cuenta las lecturas: sirve para asserting que un hit no toca
	// la base.
	getCalls int
	// setCalls cuenta las escrituras.
	setCalls int
	// deleteCalls cuenta los borrados.
	deleteCalls int
	// getErr simula una falla del cache (por ejemplo, Redis caído).
	getErr error
}

func newCacheFalso() *cacheFalso {
	return &cacheFalso{values: map[string][]byte{}}
}

func (c *cacheFalso) Get(_ context.Context, key string) ([]byte, error) {
	c.getCalls++

	if c.getErr != nil {
		return nil, c.getErr
	}

	value, ok := c.values[key]
	if !ok {
		return nil, ErrMiss
	}

	return value, nil
}

func (c *cacheFalso) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.setCalls++
	c.values[key] = value

	return nil
}

func (c *cacheFalso) Delete(_ context.Context, keys ...string) error {
	c.deleteCalls++

	for _, key := range keys {
		delete(c.values, key)
	}

	return nil
}

func (c *cacheFalso) Incr(_ context.Context, key string) (int64, error) {
	c.setCalls++

	current := int64(0)
	if value, ok := c.values[key]; ok {
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return 0, err
		}
		current = parsed
	}

	current++

	// Redis guarda el contador como el texto decimal del número.
	c.values[key] = []byte(strconv.FormatInt(current, 10))

	return current, nil
}

func TestGetOrLoad(t *testing.T) {
	t.Run("carga de la base y guarda el valor cuando el cache está vacío", func(t *testing.T) {
		c := newCacheFalso()

		loads := 0
		got, err := GetOrLoad(context.Background(), c, "clave", time.Minute,
			func(context.Context) (string, error) {
				loads++

				return "valor", nil
			})

		require.NoError(t, err)
		assert.Equal(t, "valor", got)
		assert.Equal(t, 1, loads)
		assert.Equal(t, 1, c.setCalls, "el valor cargado tiene que quedar cacheado")
	})

	t.Run("devuelve lo cacheado sin volver a la base", func(t *testing.T) {
		c := newCacheFalso()

		loads := 0
		load := func(context.Context) (string, error) {
			loads++

			return "valor", nil
		}

		_, err := GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.NoError(t, err)

		got, err := GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.NoError(t, err)

		assert.Equal(t, "valor", got)
		assert.Equal(t, 1, loads, "el segundo pedido tiene que salir del cache")
	})

	t.Run("no guarda el error del load", func(t *testing.T) {
		c := newCacheFalso()

		errEsperado := errors.New("boom")
		loads := 0
		load := func(context.Context) (string, error) {
			loads++

			return "", errEsperado
		}

		_, err := GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.ErrorIs(t, err, errEsperado)
		assert.Equal(t, 0, c.setCalls, "un error no se cachea")

		// El siguiente intento vuelve a la base: no quedó nada guardado.
		_, err = GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.ErrorIs(t, err, errEsperado)
		assert.Equal(t, 2, loads)
	})

	t.Run("un cache caído no rompe la operación", func(t *testing.T) {
		c := newCacheFalso()
		c.getErr = errors.New("redis caído")

		got, err := GetOrLoad(context.Background(), c, "clave", time.Minute,
			func(context.Context) (string, error) { return "valor", nil })

		require.NoError(t, err)
		assert.Equal(t, "valor", got, "si el cache falla, se sirve desde la base")
	})

	t.Run("un valor cacheado que no se puede deserializar se vuelve a pedir", func(t *testing.T) {
		c := newCacheFalso()

		// El cache guarda JSON: si lo guardado no es el tipo esperado, el
		// unmarshal falla y hay que ir a la base.
		require.NoError(t, c.Set(context.Background(), "clave", []byte("no es json"), time.Minute))

		loads := 0
		got, err := GetOrLoad(context.Background(), c, "clave", time.Minute,
			func(context.Context) (string, error) {
				loads++

				return "valor", nil
			})

		require.NoError(t, err)
		assert.Equal(t, "valor", got)
		assert.Equal(t, 1, loads)
	})

	t.Run("el valor cacheado conserva el tipo aunque sea un puntero", func(t *testing.T) {
		c := newCacheFalso()

		loads := 0
		load := func(context.Context) (*int, error) {
			loads++

			value := 42

			return &value, nil
		}

		got, err := GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 42, *got)

		got, err = GetOrLoad(context.Background(), c, "clave", time.Minute, load)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 42, *got)
		assert.Equal(t, 1, loads)
	})
}

func TestVersion(t *testing.T) {
	t.Run("devuelve 0 cuando no hay versión guardada", func(t *testing.T) {
		c := newCacheFalso()

		assert.Equal(t, "0", Version(context.Background(), c, "ingresos"))
	})

	t.Run("devuelve la versión guardada", func(t *testing.T) {
		c := newCacheFalso()
		require.NoError(t, c.Set(context.Background(), "ver:ingresos", []byte("7"), time.Minute))

		assert.Equal(t, "7", Version(context.Background(), c, "ingresos"))
	})

	t.Run("un scope distinto tiene su propia versión", func(t *testing.T) {
		c := newCacheFalso()
		require.NoError(t, c.Set(context.Background(), "ver:ingresos", []byte("7"), time.Minute))

		assert.Equal(t, "0", Version(context.Background(), c, "gastos"))
	})
}

func TestBump(t *testing.T) {
	t.Run("incrementa el contador del scope", func(t *testing.T) {
		c := newCacheFalso()

		Bump(context.Background(), c, "ingresos")
		Bump(context.Background(), c, "ingresos")

		count, err := c.Get(context.Background(), "ver:ingresos")
		require.NoError(t, err)
		assert.Equal(t, []byte("2"), count)
	})

	t.Run("el Bump cambia lo que devuelve la Version", func(t *testing.T) {
		// Es lo que hace útil al par Version/Bump: si el Bump incrementara
		// una clave distinta de la que lee la Version, la invalidación no
		// serviría de nada.
		c := newCacheFalso()

		assert.Equal(t, "0", Version(context.Background(), c, "ingresos"))

		Bump(context.Background(), c, "ingresos")

		assert.Equal(t, "1", Version(context.Background(), c, "ingresos"))

		Bump(context.Background(), c, "ingresos")

		assert.Equal(t, "2", Version(context.Background(), c, "ingresos"))
	})

	t.Run("cada scope tiene su propia versión", func(t *testing.T) {
		c := newCacheFalso()

		Bump(context.Background(), c, "ingresos")

		assert.Equal(t, "1", Version(context.Background(), c, "ingresos"))
		assert.Equal(t, "0", Version(context.Background(), c, "gastos"),
			"el Bump de un scope no tiene que tocar el otro")
	})
}

func TestNoop(t *testing.T) {
	t.Run("siempre misses y no guarda nada", func(t *testing.T) {
		c := Noop{}

		value, err := c.Get(context.Background(), "clave")
		require.ErrorIs(t, err, ErrMiss)
		assert.Nil(t, value)

		require.NoError(t, c.Set(context.Background(), "clave", []byte("valor"), time.Minute))

		value, err = c.Get(context.Background(), "clave")
		require.ErrorIs(t, err, ErrMiss)
		assert.Nil(t, value)
	})

	t.Run("las operaciones de escritura no fallan", func(t *testing.T) {
		c := Noop{}

		require.NoError(t, c.Delete(context.Background(), "clave"))

		count, err := c.Incr(context.Background(), "clave")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
