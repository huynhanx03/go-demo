package mapper

import (
	"search-radius/internal/core/dto"
	"search-radius/internal/core/entity"
	"time"
)

func ToShopResponse(s *entity.Shop) *dto.ShopResponse {
	return &dto.ShopResponse{
		ID:   s.ID,
		Name: s.Name,
		Lat:  s.Lat,
		Lng:  s.Lng,
	}
}

func ToShopEntityFromReq(req *dto.CreateShopRequest) *entity.Shop {
	return &entity.Shop{
		Name:      req.Name,
		Lat:       req.Lat,
		Lng:       req.Lng,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
