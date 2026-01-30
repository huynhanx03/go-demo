package di

import (
	"log"
	"path/filepath"

	"github.com/huynhanx03/go-demo/promotion-file/internal/core"
	"github.com/huynhanx03/go-demo/promotion-file/internal/infrastructure"
	"github.com/huynhanx03/go-demo/promotion-file/pkg/constants"
)

// SetupDependencies initializes the entire dependency graph.
func SetupDependencies(dataDir string) *Container {
	return &Container{
		PromotionContainer: InitPromotionDependencies(dataDir),
	}
}

// InitPromotionDependencies initializes the Promotion module dependencies.
func InitPromotionDependencies(dataDir string) *PromotionContainer {
	repo, err := infrastructure.InitIntersectionRepository(
		filepath.Join(dataDir, constants.FileCampaign),
		filepath.Join(dataDir, constants.FileMember),
	)
	if err != nil {
		log.Fatalf("Failed to init repository: %v", err)
	}

	return &PromotionContainer{
		Service:    core.NewPromotionService(repo),
		Repository: repo,
	}
}
