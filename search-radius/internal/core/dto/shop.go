package dto

type CreateShopRequest struct {
	Name string  `json:"name" validate:"required"`
	Lat  float64 `json:"lat" validate:"required"`
	Lng  float64 `json:"lng" validate:"required"`
}

type UpdateShopRequest struct {
	ID   int      `json:"-" uri:"id"`
	Name *string  `json:"name"`
	Lat  *float64 `json:"lat"`
	Lng  *float64 `json:"lng"`
}

type GetShopRequest struct {
	ID int `uri:"id" validate:"required"`
}

type DeleteShopRequest struct {
	ID int `uri:"id" validate:"required"`
}

type ShopResponse struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type SearchShopByRadiusRequest struct {
	Lat    float64 `form:"lat" json:"lat" validate:"required"`
	Lng    float64 `form:"lng" json:"lng" validate:"required"`
	Radius float64 `form:"radius" json:"radius" validate:"required,gt=0"`
}
