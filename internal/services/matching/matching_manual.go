package matching

import (
	"context"
	"log"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// ManualMatcher finds all matches but doesn't auto-assign (for admin review)
type ManualMatcher struct {
	discovery *ServiceDiscovery
}

// NewManualMatcher creates a new manual matcher
func NewManualMatcher() *ManualMatcher {
	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	return &ManualMatcher{
		discovery: NewServiceDiscovery(repo),
	}
}

// Assign does not assign - order remains unassigned for manual selection
func (m *ManualMatcher) Assign(ctx context.Context, order *models.Order, options models.PrintOrderOptions) error {
	// Don't assign - order remains unassigned
	order.PrintShopID = ""

	log.Printf("📋 Manual mode: Order %s requires manual shop assignment", order.OrderID)
	return nil
}

// GetMatches returns all matching shops for an order (for admin review)
func (m *ManualMatcher) GetMatches(ctx context.Context, order *models.Order, options models.PrintOrderOptions) ([]models.ShopMatch, error) {
	matches, err := m.discovery.FindMatchingShops(ctx, options)
	if err != nil {
		log.Printf("❌ Failed to find matching shops: %v", err)
		return nil, err
	}

	// Sort by match score for admin review
	// In manual mode, we can show all options sorted by various criteria
	// For now, sort by price (cheapest first)
	for i := range matches {
		// Calculate basic score for sorting
		matches[i].MatchScore = 100.0 / (1.0 + matches[i].TotalPrice/100.0)
	}

	return matches, nil
}
