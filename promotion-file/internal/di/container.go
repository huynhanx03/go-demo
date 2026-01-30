package di

import "github.com/huynhanx03/go-demo/promotion-file/internal/ports"

// PromotionContainer holds dependencies for the Promotion module.
type PromotionContainer struct {
	Service    ports.IPromotionService
	Repository ports.IPromotionRepository
}

// Container defines the global dependency container.
type Container struct {
	PromotionContainer *PromotionContainer
}
