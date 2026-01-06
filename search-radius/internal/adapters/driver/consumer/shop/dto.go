package shop

import (
	"search-radius/internal/core/entity"
	"time"
)

// CDCShop represents the Debezium JSON structure for a shop row
type CDCShop struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *CDCShop) ToEntity() *entity.Shop {
	if c == nil {
		return nil
	}
	return &entity.Shop{
		ID:        c.ID,
		Name:      c.Name,
		Lat:       c.Lat,
		Lng:       c.Lng,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
