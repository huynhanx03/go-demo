package di

import (
	"context"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"search-radius/global"
	"search-radius/internal/adapters/driven/cache"
	"search-radius/internal/adapters/driven/db"
	"search-radius/internal/adapters/driven/db/ent/generate"
	shopconsumer "search-radius/internal/adapters/driver/consumer/shop"
	driverHttp "search-radius/internal/adapters/driver/http"
	"search-radius/internal/core/service"
	"search-radius/internal/ports"
	"search-radius/pkg/database/ent"
	"search-radius/pkg/mq/kafka"
)

func InitShopDependencies() (
	ports.ShopRepository,
	ports.ShopService,
	driverHttp.ShopHandler,
	ports.ShopConsumer,
) {
	shopDriver, err := ent.NewDriver(global.Config.Database)
	if err != nil {
		log.Fatalf("failed opening connection to ent: %v", err)
	}

	shopClient := generate.NewClient(generate.Driver(shopDriver)).Debug()
	if err := shopClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// Repositories
	shopRepository := db.NewShop(shopClient)
	shopCache := cache.NewShopCache(global.Redis)

	// Service
	shopService := service.NewShop(shopRepository, shopCache)

	// Handler
	shopHandler := driverHttp.NewShop(shopService)

	// Consumer
	kafkaCfg := &kafka.Config{
		Brokers:  global.Config.Kafka.Brokers,
		ClientID: "shop-cdc",
		ProducerInfo: kafka.ProducerConfig{
			FlushFrequency:  global.Config.Kafka.FlushFrequency,
			FlushBytes:      global.Config.Kafka.FlushBytes,
			MaxMessageBytes: global.Config.Kafka.MaxMessageBytes,
			MaxRetries:      global.Config.Kafka.MaxRetries,
			RetryBackoff:    global.Config.Kafka.RetryBackoff,
			ReturnSuccesses: true,
		},
		ConsumerInfo: kafka.ConsumerConfig{
			SessionTimeout:    global.Config.Kafka.Timeout * 1000,
			MaxProcessingTime: global.Config.Kafka.MaxProcessingTime,
		},
	}

	ShopConsumer, err := shopconsumer.NewCDCConsumer(kafkaCfg, shopService)
	if err != nil {
		global.Logger.Fatal("failed to create shop cdc consumer", zap.Error(err))
	}

	return shopRepository, shopService, shopHandler, ShopConsumer
}
