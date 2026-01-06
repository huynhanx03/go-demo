package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"search-radius/internal/core/dto"
	"search-radius/internal/core/entity"
	"search-radius/internal/core/mapper"
	"search-radius/internal/ports"
	"search-radius/pkg/cdc"
	"search-radius/pkg/common/apperr"
	"search-radius/pkg/common/http/response"
	d "search-radius/pkg/dto"
)

type shopService struct {
	shopRepo  ports.ShopRepository
	shopCache ports.ShopCacheRepository
}

func NewShop(
	shopRepo ports.ShopRepository,
	shopCache ports.ShopCacheRepository,
) ports.ShopService {
	return &shopService{
		shopRepo:  shopRepo,
		shopCache: shopCache,
	}
}

// Find shops with pagination
func (s *shopService) Find(ctx context.Context, opts *d.QueryOptions) (*d.Paginated[*dto.ShopResponse], error) {
	shops, err := s.shopRepo.Find(ctx, opts)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to find shops", http.StatusInternalServerError)
	}

	if shops.Records == nil {
		return &d.Paginated[*dto.ShopResponse]{
			Records:    &[]*dto.ShopResponse{},
			Pagination: shops.Pagination,
		}, nil
	}

	shopEntities := *shops.Records
	shopResponses := make([]*dto.ShopResponse, len(shopEntities))
	for i, shop := range shopEntities {
		shopResponses[i] = mapper.ToShopResponse(shop)
	}

	return &d.Paginated[*dto.ShopResponse]{
		Records:    &shopResponses,
		Pagination: shops.Pagination,
	}, nil
}

// Get shop by id
func (s *shopService) Get(ctx context.Context, id int) (*dto.ShopResponse, error) {
	shop, err := s.shopRepo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to get shop", http.StatusInternalServerError)
	}

	return mapper.ToShopResponse(shop), nil
}

// Create shop
func (s *shopService) Create(ctx context.Context, req *dto.CreateShopRequest) (*dto.ShopResponse, error) {
	shop := mapper.ToShopEntityFromReq(req)

	if err := s.shopRepo.Create(ctx, shop); err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to create shop", http.StatusInternalServerError)
	}

	return mapper.ToShopResponse(shop), nil
}

// Update shop by id
func (s *shopService) Update(ctx context.Context, id int, req *dto.UpdateShopRequest) (*dto.ShopResponse, error) {
	shop, err := s.shopRepo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to get shop", http.StatusInternalServerError)
	}

	if req.Name != nil {
		shop.Name = *req.Name
	}
	if req.Lat != nil {
		shop.Lat = *req.Lat
	}
	if req.Lng != nil {
		shop.Lng = *req.Lng
	}

	if err := s.shopRepo.Update(ctx, id, shop); err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to update shop", http.StatusInternalServerError)
	}

	return mapper.ToShopResponse(shop), nil
}

// Delete shop by id
func (s *shopService) Delete(ctx context.Context, id int) error {
	exists, err := s.shopRepo.Exists(ctx, id)
	if err != nil {
		return apperr.Wrap(err, response.CodeDatabaseError, "failed to check shop exists", http.StatusInternalServerError)
	}

	if !exists {
		return apperr.New(response.CodeNotFound, "shop not found", http.StatusNotFound, nil)
	}

	if err := s.shopRepo.Delete(ctx, id); err != nil {
		return apperr.Wrap(err, response.CodeDatabaseError, "failed to delete shop", http.StatusInternalServerError)
	}

	return nil
}

// UpdateLocation updates shop location
func (s *shopService) UpdateLocation(ctx context.Context, id int, lat, lng float64) error {
	shop, err := s.shopRepo.Get(ctx, id)
	if err != nil {
		return apperr.Wrap(err, response.CodeDatabaseError, "failed to get shop", http.StatusInternalServerError)
	}

	shop.Lat = lat
	shop.Lng = lng

	if err := s.shopRepo.Update(ctx, id, shop); err != nil {
		return apperr.Wrap(err, response.CodeDatabaseError, "failed to update shop location", http.StatusInternalServerError)
	}

	return nil
}

// HandleShopBatchChange handles batch CDC events for shops
func (s *shopService) HandleShopBatchChange(ctx context.Context, batch []*cdc.DebeziumPayload[entity.Shop]) error {
	var (
		shopsToSave []*entity.Shop
		idsToRemove []int
	)

	for _, payload := range batch {
		switch payload.Op {
		case cdc.OpCreate, cdc.OpUpdate:
			if payload.After != nil {
				shopsToSave = append(shopsToSave, payload.After)
			}
		case cdc.OpDelete:
			if payload.Before != nil {
				idsToRemove = append(idsToRemove, payload.Before.ID)
			}
		}
	}

	// Update/Add locations
	if len(shopsToSave) > 0 {
		if err := s.shopCache.BatchSaveGeo(ctx, shopsToSave); err != nil {
			return apperr.Wrap(err, response.CodeInternalServer, "failed to batch save shop geo", http.StatusInternalServerError)
		}
	}

	// Remove deleted locations
	if len(idsToRemove) > 0 {
		if err := s.shopCache.BatchRemoveGeo(ctx, idsToRemove); err != nil {
			return apperr.Wrap(err, response.CodeInternalServer, "failed to batch remove shop geo", http.StatusInternalServerError)
		}
	}

	return nil
}

// SearchByRadius searches shops by radius
func (s *shopService) SearchByRadius(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error) {
	shops, err := s.shopRepo.SearchRadius(ctx, req.Lat, req.Lng, req.Radius)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to search shops by radius", http.StatusInternalServerError)
	}

	responses := make([]*dto.ShopResponse, len(shops))
	for i, shop := range shops {
		responses[i] = mapper.ToShopResponse(shop)
	}
	return responses, nil
}

// SearchByRadiusFast searches shops by radius using Redis Geo
func (s *shopService) SearchByRadiusFast(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error) {
	ids, err := s.shopCache.SearchRadius(ctx, req.Lat, req.Lng, req.Radius)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeInternalServer, "failed to search shops by radius (redis)", http.StatusInternalServerError)
	}

	if len(ids) == 0 {
		return []*dto.ShopResponse{}, nil
	}

	shops, err := s.shopRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperr.Wrap(err, response.CodeDatabaseError, "failed to get shops by ids", http.StatusInternalServerError)
	}

	responses := make([]*dto.ShopResponse, len(shops))
	for i, shop := range shops {
		responses[i] = mapper.ToShopResponse(shop)
	}

	return responses, nil
}

func (s *shopService) InitData(ctx context.Context, count int) error {
	const batchSize = 1000
	var shops []*entity.Shop

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Vietnam bounding box
	minLat, maxLat := 8.5, 23.5
	minLng, maxLng := 102.0, 109.5

	for i := 0; i < count; i++ {
		lat := minLat + rng.Float64()*(maxLat-minLat)
		lng := minLng + rng.Float64()*(maxLng-minLng)

		shops = append(shops, &entity.Shop{
			Name: fmt.Sprintf("Shop %d", i+1),
			Lat:  lat,
			Lng:  lng,
		})

		if len(shops) >= batchSize {
			if err := s.shopRepo.CreateBatch(ctx, shops); err != nil {
				return err
			}
			shops = shops[:0]
		}
	}

	if len(shops) > 0 {
		return s.shopRepo.CreateBatch(ctx, shops)
	}
	return nil
}
