package http

import (
	"context"

	"search-radius/internal/core/dto"
	"search-radius/internal/ports"
	"search-radius/pkg/common/http/handler"
	d "search-radius/pkg/dto"
)

type ShopHandler interface {
	Find(ctx context.Context, req *d.QueryOptions) (*d.Paginated[*dto.ShopResponse], error)
	Get(ctx context.Context, req *dto.GetShopRequest) (*dto.ShopResponse, error)
	Create(ctx context.Context, req *dto.CreateShopRequest) (*dto.ShopResponse, error)
	Update(ctx context.Context, req *dto.UpdateShopRequest) (*dto.ShopResponse, error)
	Delete(ctx context.Context, req *dto.DeleteShopRequest) (*dto.ShopResponse, error)
	SearchByRadius(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error)
	SearchByRadiusFast(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error)
}

type shopHandler struct {
	handler.BaseHandler
	shopService ports.ShopService
}

func NewShop(shopService ports.ShopService) ShopHandler {
	return &shopHandler{
		shopService: shopService,
	}
}

// Find shops
func (h *shopHandler) Find(ctx context.Context, req *d.QueryOptions) (*d.Paginated[*dto.ShopResponse], error) {
	return h.shopService.Find(ctx, req)
}

// Get shop by id
func (h *shopHandler) Get(ctx context.Context, req *dto.GetShopRequest) (*dto.ShopResponse, error) {
	return h.shopService.Get(ctx, req.ID)
}

// Create shop
func (h *shopHandler) Create(ctx context.Context, req *dto.CreateShopRequest) (*dto.ShopResponse, error) {
	return h.shopService.Create(ctx, req)
}

// Update shop
func (h *shopHandler) Update(ctx context.Context, req *dto.UpdateShopRequest) (*dto.ShopResponse, error) {
	return h.shopService.Update(ctx, req.ID, req)
}

// Delete shop
func (h *shopHandler) Delete(ctx context.Context, req *dto.DeleteShopRequest) (*dto.ShopResponse, error) {
	return nil, h.shopService.Delete(ctx, req.ID)
}

// SearchByRadius searches shops by radius
func (h *shopHandler) SearchByRadius(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error) {
	return h.shopService.SearchByRadius(ctx, req)
}

// SearchByRadiusFast searches shops by radius fast
func (h *shopHandler) SearchByRadiusFast(ctx context.Context, req *dto.SearchShopByRadiusRequest) ([]*dto.ShopResponse, error) {
	return h.shopService.SearchByRadiusFast(ctx, req)
}
