package di

func SetupDependencies() {
	shopRepository, shopService, shopHandler, shopConsumer := InitShopDependencies()

	GlobalContainer = &Container{
		shopRepository: shopRepository,
		ShopService:    shopService,
		ShopHandler:    shopHandler,
		ShopConsumer:   shopConsumer,
	}
}
