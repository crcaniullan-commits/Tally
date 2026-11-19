package users

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/cache"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

const ttlUsuario = 60 * time.Second

type buscadorPorID interface {
	ObtenerPorID(context.Context, uuid.UUID) (*model.User, error)
}

type BuscadorUsuarioConCache struct {
	siguiente buscadorPorID
	cache     cache.Cache
}

func NuevoBuscadorUsuarioConCache(s buscadorPorID, c cache.Cache) *BuscadorUsuarioConCache {
	return &BuscadorUsuarioConCache{siguiente: s, cache: c}
}

func claveUsuario(id uuid.UUID) string { return "user:" + id.String() }

func (b *BuscadorUsuarioConCache) ObtenerPorID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return cache.GetOrLoad(ctx, b.cache, claveUsuario(id), ttlUsuario,
		func(ctx context.Context) (*model.User, error) {
			return b.siguiente.ObtenerPorID(ctx, id)
		})
}

func (b *BuscadorUsuarioConCache) Invalidar(ctx context.Context, id uuid.UUID) {
	_ = b.cache.Delete(ctx, claveUsuario(id))
}
