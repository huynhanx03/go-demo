package ports

// IPromotionService defines the interface for business logic.
type IPromotionService interface {
	IsEligible(code string) bool
}
