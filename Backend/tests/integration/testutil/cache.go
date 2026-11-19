package testutil

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/cache"
)

// cacheMemoria es un cache.Cache en memoria, para que la suite pueda
// ejercitar el camino cacheado (el de internal/cache) sin depender de un
// Redis.
//
// No es un Redis: no hay TTL real (los valores viven hasta que se borran), ni
// errores de red. Lo que importa es que guarde y devuelva lo mismo que
// guardaría el cache real, así que un bug de serialización (por ejemplo, un
// campo con json:"-") se reproduce acá.
type cacheMemoria struct {
	mu     sync.Mutex
	values map[string][]byte
}

func newCacheMemoria() *cacheMemoria {
	return &cacheMemoria{values: map[string][]byte{}}
}

func (c *cacheMemoria) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.values[key]
	if !ok {
		return nil, cache.ErrMiss
	}

	return value, nil
}

func (c *cacheMemoria) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[key] = value

	return nil
}

func (c *cacheMemoria) Delete(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.values, key)
	}

	return nil
}

// Incr incrementa el contador de la clave y devuelve el valor nuevo. Es lo que
// usa cache.Bump para invalidar listados completos.
func (c *cacheMemoria) Incr(_ context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := int64(0)
	if value, ok := c.values[key]; ok {
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return 0, err
		}
		current = parsed
	}

	current++

	c.values[key] = []byte(strconv.FormatInt(current, 10))

	return current, nil
}
