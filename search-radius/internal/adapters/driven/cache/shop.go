package cache

import (
	"context"
	"strconv"

	"search-radius/internal/core/entity"
	"search-radius/internal/ports"
	"search-radius/pkg/common/cache"
)

const (
	KeyShopGeo = "shops:geo"
)

type shopCache struct {
	redis cache.CacheEngine
}

func NewShopCache(redis cache.CacheEngine) ports.ShopCacheRepository {
	return &shopCache{
		redis: redis,
	}
}

// BatchSaveGeo updates shop locations in bulk using Redis Geo
func (c *shopCache) BatchSaveGeo(ctx context.Context, shops []*entity.Shop) error {
	if len(shops) == 0 {
		return nil
	}

	locations := make([]*cache.GeoLocation, len(shops))
	for i, shop := range shops {
		locations[i] = &cache.GeoLocation{
			Member:    strconv.Itoa(shop.ID),
			Longitude: shop.Lng,
			Latitude:  shop.Lat,
		}
	}

	return c.redis.GeoAdd(ctx, KeyShopGeo, locations...)
}

// BatchRemoveGeo removes shops from geo index
func (c *shopCache) BatchRemoveGeo(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	members := make([]string, len(ids))
	for i, id := range ids {
		members[i] = strconv.Itoa(id)
	}

	return c.redis.GeoRemove(ctx, KeyShopGeo, members...)
}

// SearchRadius finds shop IDs within a radius
func (c *shopCache) SearchRadius(ctx context.Context, lat, lng, radius float64) ([]int, error) {
	locations, err := c.redis.GeoRadius(ctx, KeyShopGeo, lng, lat, radius, "km")
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(locations))
	for _, loc := range locations {
		id, err := strconv.Atoi(loc.Member)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}
