package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrMiss = errors.New("cache: miss")

type Cache interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
	Delete(context.Context, ...string) error
	Incr(context.Context, string) (int64, error)
}

// GetOrLoad implenta cache-aside, los errores de load (incluido "no encontrado")
// no se guardan enn caché cualquier falla de Redis se trata como un fallo de caché.
func GetOrLoad[T any](ctx context.Context, c Cache, key string, ttl time.Duration, load func(context.Context) (T, error)) (T, error) {
	var zero T

	if b, err := c.Get(ctx, key); err == nil {
		var v T
		if json.Unmarshal(b, &v) == nil {
			return v, nil
		}
	}

	v, err := load(ctx)
	if err != nil {
		return zero, err
	}

	if b, err := json.Marshal(v); err == nil {
		_ = c.Set(ctx, key, b, ttl)
	}

	return v, nil
}

// claveVersion es la clave del contador de versión de un scope.
//
// Version y Bump tienen que usar LA MISMA clave: si no, un Bump incrementaría
// un contador que Version nunca lee y la invalidación no serviría de nada.
func claveVersion(scope string) string {
	return "ver:" + scope
}

// Version Y Bump sirven para invalidar listas completas.
//
// Version devuelve la versión actual del scope ("0" si todavía no hubo
// ningún Bump). Bump la incrementa, así que la próxima lista que se arme con
// esa versión no usa la entrada vieja.
func Version(ctx context.Context, c Cache, scope string) string {
	b, err := c.Get(ctx, claveVersion(scope))

	if err != nil {
		return "0"
	}
	return string(b)
}

func Bump(ctx context.Context, c Cache, scope string) {
	_, _ = c.Incr(ctx, claveVersion(scope))
}

// Noop no guarda nada: el backend funciona igual sin Redis.
type Noop struct{}

func (Noop) Get(context.Context, string) ([]byte, error)              { return nil, ErrMiss }
func (Noop) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (Noop) Delete(context.Context, ...string) error                  { return nil }
func (Noop) Incr(context.Context, string) (int64, error)              { return 0, nil }
