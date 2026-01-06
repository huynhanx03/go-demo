package mapper

import (
	"search-radius/internal/adapters/driven/db/ent/generate"
	"search-radius/internal/core/entity"
)

func ToShopEntity(m *generate.Shop) *entity.Shop {
	if m == nil {
		return nil
	}
	return &entity.Shop{
		ID:        m.ID,
		Name:      m.Name,
		Lat:       m.Lat,
		Lng:       m.Lng,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToShopModel(e *entity.Shop) *generate.Shop {
	if e == nil {
		return nil
	}
	return &generate.Shop{
		ID:        e.ID,
		Name:      e.Name,
		Lat:       e.Lat,
		Lng:       e.Lng,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
