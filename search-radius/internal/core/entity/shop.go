package entity

import "time"

type Shop struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
