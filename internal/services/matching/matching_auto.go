package matching

import (
	"context"
	"log"
	"sort"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// AutoMatcher automatically assigns orders to the cheapest available shop
type AutoMatcher struct {
	discovery *ServiceDiscovery
}

// NewAutoMatcher creates a new auto matcher
func NewAutoMatcher() *AutoMatcher {
	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	return &AutoMatcher{
		discovery: NewServiceDiscovery(repo),
	}
}

// Assign automatically assigns the order to the cheapest matching shop
func (m *AutoMatcher) Assign(ctx context.Context, order *models.Order, options models.PrintOrderOptions) error {
	// Find all matching shops
	matches, err := m.discovery.FindMatchingShops(ctx, options)
	if err != nil {
		log.Printf("❌ Failed to find matching shops: %v", err)
		return err
	}

	if len(matches) == 0 {
		log.Printf("⚠️ No matching shops found for order %s", order.OrderID)
		return nil // Order remains unassigned
	}

	// Sort by price (cheapest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].TotalPrice < matches[j].TotalPrice
	})

	// Assign to cheapest shop
	bestMatch := matches[0]
	order.PrintShopID = bestMatch.ShopID

	log.Printf("✅ Auto-assigned order %s to shop %s (price: %.2f)", 
		order.OrderID, bestMatch.ShopName, bestMatch.TotalPrice)

	return nil
}
