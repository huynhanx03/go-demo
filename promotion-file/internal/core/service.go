package core

import "github.com/huynhanx03/go-demo/promotion-file/internal/ports"

type promotionService struct {
	repo ports.IPromotionRepository
}

// NewPromotionService creates a new instance of IPromotionService.
func NewPromotionService(repo ports.IPromotionRepository) ports.IPromotionService {
	return &promotionService{
		repo: repo,
	}
}

func (s *promotionService) IsEligible(code string) bool {
	return s.repo.Contains(code)
}
