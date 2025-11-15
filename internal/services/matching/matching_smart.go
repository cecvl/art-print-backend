package matching

import "github.com/cecvl/art-print-backend/internal/models"

// Smart matching stubs for future logic
type SmartMatcher struct{}

func NewSmartMatcher() *SmartMatcher {
	return &SmartMatcher{}
}

func (m *SmartMatcher) Assign(order *models.Order) error {
	// add advanced logic
	order.PrintShopID = "smart-selected-shop"
	return nil
}
