package ports

// IPromotionRepository defines the interface for data access.
type IPromotionRepository interface {
	Contains(code string) bool
}
