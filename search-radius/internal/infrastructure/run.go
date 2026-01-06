package infrastructure

import (
	"context"
	"search-radius/global"
	"search-radius/internal/di"

	"go.uber.org/zap"
)

func Run() error {
	LoadConfig()
	SetupLogger()
	SetupRedis()
	di.SetupDependencies()
	http := NewHTTPServer()

	ctx := context.Background()

	if err := di.GlobalContainer.ShopConsumer.Start(ctx); err != nil {
		global.Logger.Error("Shop CDC Consumer failed", zap.Error(err))
	}

	if err := di.GlobalContainer.ShopService.InitData(ctx, 1000000); err != nil {
		global.Logger.Error("failed to init shop data", zap.Error(err))
	} else {
		global.Logger.Info("shop data check/init completed")
	}

	return http.Run()
}
