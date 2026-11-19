package empresas

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/cache"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cacheFalso es un cache.Cache en memoria: guarda los bytes tal cual, sin
// interpretar, como haría Redis.
type cacheFalso struct {
	values map[string][]byte
}

func newCacheFalso() *cacheFalso {
	return &cacheFalso{values: map[string][]byte{}}
}

func (c *cacheFalso) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := c.values[key]
	if !ok {
		return nil, cache.ErrMiss
	}

	return value, nil
}

func (c *cacheFalso) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.values[key] = value

	return nil
}

func (c *cacheFalso) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.values, key)
	}

	return nil
}

func (c *cacheFalso) Incr(_ context.Context, key string) (int64, error) {
	return 0, nil
}

// newTestEmpresaPtr devuelve la empresa de prueba como puntero, que es lo que
// devuelve el store.
func newTestEmpresaPtr() *model.Empresa {
	empresa := newTestEmpresa()

	return &empresa
}

// almacenBuscadorFake cuenta las llamadas a ObtenerPorID: sirve para
// asserting que un hit del cache no vuelve a la base.
type almacenBuscadorFake struct {
	empresa *model.Empresa
	err     error
	calls   int
}

func (a *almacenBuscadorFake) ObtenerPorID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*model.Empresa, error) {
	a.calls++

	if a.err != nil {
		return nil, a.err
	}

	return a.empresa, nil
}

func TestBuscadorEmpresaConCache_ObtenerPorID(t *testing.T) {
	t.Run("carga de la base la primera vez", func(t *testing.T) {
		store := &almacenBuscadorFake{empresa: newTestEmpresaPtr()}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		empresa, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)

		require.NoError(t, err)
		assert.Equal(t, testEmpresaID, empresa.ID)
		assert.Equal(t, 1, store.calls)
		assert.Len(t, c.values, 1, "la empresa tiene que quedar cacheada")
	})

	t.Run("la segunda vez sale del cache", func(t *testing.T) {
		store := &almacenBuscadorFake{empresa: newTestEmpresaPtr()}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)

		empresa, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)

		assert.Equal(t, testEmpresaID, empresa.ID)
		assert.Equal(t, 1, store.calls, "el segundo pedido tiene que salir del cache")
	})

	t.Run("la clave incluye el usuario: dos usuarios no comparten entrada", func(t *testing.T) {
		store := &almacenBuscadorFake{empresa: newTestEmpresaPtr()}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		otro := uuid.New()

		_, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)

		_, err = buscador.ObtenerPorID(context.Background(), testEmpresaID, otro)
		require.NoError(t, err)

		assert.Equal(t, 2, store.calls, "el otro usuario no puede leer la entrada del primero")
		assert.Len(t, c.values, 2)
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		errEsperado := errors.New("boom")
		store := &almacenBuscadorFake{err: errEsperado}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)

		require.ErrorIs(t, err, errEsperado)
		assert.Empty(t, c.values, "un error no se cachea")
	})
}

func TestBuscadorEmpresaConCache_Invalidar(t *testing.T) {
	t.Run("borra la entrada de la empresa", func(t *testing.T) {
		store := &almacenBuscadorFake{empresa: newTestEmpresaPtr()}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)
		require.Len(t, c.values, 1)

		buscador.Invalidar(context.Background(), testUserID, testEmpresaID)

		assert.Empty(t, c.values, "la entrada tiene que desaparecer del cache")

		// El siguiente pedido vuelve a la base.
		_, err = buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)
		assert.Equal(t, 2, store.calls)
	})

	t.Run("invalidar una empresa no borra la de otro usuario", func(t *testing.T) {
		store := &almacenBuscadorFake{empresa: newTestEmpresaPtr()}
		c := newCacheFalso()
		buscador := NuevoBuscadorEmpresaConCache(store, c)

		otro := uuid.New()

		_, err := buscador.ObtenerPorID(context.Background(), testEmpresaID, testUserID)
		require.NoError(t, err)

		_, err = buscador.ObtenerPorID(context.Background(), testEmpresaID, otro)
		require.NoError(t, err)

		buscador.Invalidar(context.Background(), testUserID, testEmpresaID)

		assert.Len(t, c.values, 1, "la entrada del otro usuario tiene que sobrevivir")
	})
}
