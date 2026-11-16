package users

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
	values      map[string][]byte
	deleteCalls int
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
	c.deleteCalls++

	for _, key := range keys {
		delete(c.values, key)
	}

	return nil
}

func (c *cacheFalso) Incr(_ context.Context, key string) (int64, error) {
	return 0, nil
}

// almacenBuscadorFake cuenta las llamadas a ObtenerPorID: sirve para
// asserting que un hit del cache no vuelve a la base.
type almacenBuscadorFake struct {
	user  *model.User
	err   error
	calls int
}

func (a *almacenBuscadorFake) ObtenerPorID(_ context.Context, _ uuid.UUID) (*model.User, error) {
	a.calls++

	if a.err != nil {
		return nil, a.err
	}

	return a.user, nil
}

func TestBuscadorUsuarioConCache_ObtenerPorID(t *testing.T) {
	t.Run("carga de la base la primera vez", func(t *testing.T) {
		store := &almacenBuscadorFake{user: newTestUser()}
		c := newCacheFalso()
		buscador := NuevoBuscadorUsuarioConCache(store, c)

		user, err := buscador.ObtenerPorID(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, 1, store.calls)
		assert.Len(t, c.values, 1, "el usuario tiene que quedar cacheado")
	})

	t.Run("la segunda vez sale del cache", func(t *testing.T) {
		store := &almacenBuscadorFake{user: newTestUser()}
		c := newCacheFalso()
		buscador := NuevoBuscadorUsuarioConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testUserID)
		require.NoError(t, err)

		user, err := buscador.ObtenerPorID(context.Background(), testUserID)
		require.NoError(t, err)

		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, 1, store.calls, "el segundo pedido tiene que salir del cache")
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		errEsperado := errors.New("boom")
		store := &almacenBuscadorFake{err: errEsperado}
		c := newCacheFalso()
		buscador := NuevoBuscadorUsuarioConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testUserID)

		require.ErrorIs(t, err, errEsperado)
		assert.Empty(t, c.values, "un error no se cachea")
	})
}

func TestBuscadorUsuarioConCache_Invalidar(t *testing.T) {
	t.Run("borra la entrada del usuario", func(t *testing.T) {
		store := &almacenBuscadorFake{user: newTestUser()}
		c := newCacheFalso()
		buscador := NuevoBuscadorUsuarioConCache(store, c)

		_, err := buscador.ObtenerPorID(context.Background(), testUserID)
		require.NoError(t, err)
		require.Len(t, c.values, 1)

		buscador.Invalidar(context.Background(), testUserID)

		assert.Empty(t, c.values, "la entrada tiene que desaparecer del cache")

		// El siguiente pedido vuelve a la base.
		_, err = buscador.ObtenerPorID(context.Background(), testUserID)
		require.NoError(t, err)
		assert.Equal(t, 2, store.calls)
	})
}
