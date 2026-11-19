package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct{ rdb *redis.Client }

func NewRedis(url string) (*RedisCache, error) {
	opt, err := redis.ParseURL(url)

	if err != nil {
		return nil, err
	}

	// timeouts cortos: si Redis no responde, preferimos ir a la base antes que esperar
	opt.DialTimeout = 200 * time.Millisecond
	opt.ReadTimeout = 150 * time.Millisecond
	opt.WriteTimeout = 150 * time.Millisecond

	return &RedisCache{rdb: redis.NewClient(opt)}, nil
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrMiss
	}
	return b, err
}

func (c *RedisCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, val, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *RedisCache) Close() error { return c.rdb.Close() }
