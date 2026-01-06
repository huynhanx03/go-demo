package di

import (
	"search-radius/internal/adapters/driver/http"
	"search-radius/internal/ports"
)

type Container struct {
	shopRepository ports.ShopRepository
	ShopService    ports.ShopService
	ShopHandler    http.ShopHandler
	ShopConsumer   ports.ShopConsumer
}

var GlobalContainer *Container
