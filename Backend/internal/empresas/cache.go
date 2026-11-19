package empresas

import (
	"context"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/cache"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

const ttlEmpresa = 60 * time.Second

type buscadorPorID interface {
	ObtenerPorID(context.Context, uuid.UUID, uuid.UUID) (*model.Empresa, error)
}

type BuscadorEmpresaConCache struct {
	siguiente buscadorPorID
	cache     cache.Cache
}

func NuevoBuscadorEmpresaConCache(s buscadorPorID, c cache.Cache) *BuscadorEmpresaConCache {
	return &BuscadorEmpresaConCache{siguiente: s, cache: c}
}

func claveEmpresa(userID, empresaID uuid.UUID) string {
	return "empresa:" + userID.String() + "" + empresaID.String()
}

func (b *BuscadorEmpresaConCache) ObtenerPorID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Empresa, error) {
	return cache.GetOrLoad(ctx, b.cache, claveEmpresa(userID, id), ttlEmpresa,
		func(ctx context.Context) (*model.Empresa, error) {
			return b.siguiente.ObtenerPorID(ctx, id, userID)
		})
}

func (b *BuscadorEmpresaConCache) Invalidar(ctx context.Context, userID, empresaID uuid.UUID) {
	_ = b.cache.Delete(ctx, claveEmpresa(userID, empresaID))
}
