package ports

import (
	"context"

	"search-radius/internal/core/dto"
	"search-radius/internal/core/entity"
	"search-radius/pkg/cdc"
	d "search-radius/pkg/dto"
)

type ShopRepository interface {
	Find(ctx context.Context, opts *d.QueryOptions) (*d.Paginated[*entity.Shop], error)
	Get(ctx context.Context, id int) (*entity.Shop, error)
	GetByIDs(ctx context.Context, ids []int) ([]*entity.Shop, error)
	Create(ctx context.Context, shop *entity.Shop) error
	CreateBatch(ctx context.Context, shops []*entity.Shop) error
	Update(ctx context.Context, id int, shop *entity.Shop) error
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)
	SearchRadius(ctx context.Context, lat, lng, radius float64) ([]*entity.Shop, error)
}

type ShopService interface {
	Find(ctx context.Context, opts *d.QueryOptions) (*d.Paginated[*dto.ShopResponse], error)
	Get(ctx context.Context, id int) (*dto.ShopResponse, error)
	Create(ctx context.Context, req *dto.CreateShopRequest) (*dto.ShopResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateShopRequest) (*dto.ShopResponse, error)
	Delete(ctx context.Context, id int) error
	UpdateLocation(ctx context.Context, id int, lat, lng float64) error
	HandleShopBatchChange(ctx context.Context, batch []*cdc.DebeziumPayload[entity.Shop]) error
	InitData(ctx context.Context, count int) error

	SearchByRadius(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error)
	SearchByRadiusFast(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error)
}

type ShopCacheRepository interface {
	BatchSaveGeo(ctx context.Context, shops []*entity.Shop) error
	BatchRemoveGeo(ctx context.Context, ids []int) error
	SearchRadius(ctx context.Context, lat, lng, radius float64) ([]int, error)
}

type ShopConsumer interface {
	Start(ctx context.Context) error
	Stop() error
}
