package matching

import "github.com/cecvl/art-print-backend/internal/models"

// Manual mode does not assign anything
type ManualMatcher struct{}

func NewManualMatcher() *ManualMatcher {
	return &ManualMatcher{}
}

func (m *ManualMatcher) Assign(order *models.Order) error {
	order.PrintShopID = "" // remains unassigned
	return nil
}
