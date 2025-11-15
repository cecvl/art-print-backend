package matching

import "github.com/cecvl/art-print-backend/internal/models"

// Auto mode picks the first eligible shop (later: ranking logic)
type AutoMatcher struct{}

func NewAutoMatcher() *AutoMatcher {
	return &AutoMatcher{}
}

func (m *AutoMatcher) Assign(order *models.Order) error {
	// add real logic
	order.PrintShopID = "auto-shop-1"
	return nil
}
