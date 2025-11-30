package matching

import (
	"context"
	"log"
	"math"
	"sort"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// SmartMatcher uses advanced scoring algorithm to find the best shop match
type SmartMatcher struct {
	discovery *ServiceDiscovery
}

// NewSmartMatcher creates a new smart matcher
func NewSmartMatcher() *SmartMatcher {
	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	return &SmartMatcher{
		discovery: NewServiceDiscovery(repo),
	}
}

// Assign uses smart scoring to assign order to best matching shop
func (m *SmartMatcher) Assign(ctx context.Context, order *models.Order, options models.PrintOrderOptions) error {
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

	// Get shop details for scoring
	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	shopsMap := make(map[string]*models.PrintShopProfile)
	for _, match := range matches {
		shop, err := repo.GetShopByID(ctx, match.ShopID)
		if err == nil {
			shopsMap[match.ShopID] = shop
		}
	}

	// Calculate smart scores for each match
	for i := range matches {
		shop := shopsMap[matches[i].ShopID]
		if shop != nil {
			matches[i].MatchScore = m.calculateSmartScore(shop, &matches[i], options)
		}
	}

	// Sort by match score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].MatchScore > matches[j].MatchScore
	})

	// Assign to best match
	bestMatch := matches[0]
	order.PrintShopID = bestMatch.ShopID

	log.Printf("✅ Smart-assigned order %s to shop %s (score: %.2f, price: %.2f)",
		order.OrderID, bestMatch.ShopName, bestMatch.MatchScore, bestMatch.TotalPrice)

	return nil
}

// calculateSmartScore computes a comprehensive match score
// Score components:
// - Price competitiveness: 40%
// - Shop rating: 30%
// - Delivery time: 20%
// - Service quality (technology): 10%
func (m *SmartMatcher) calculateSmartScore(shop *models.PrintShopProfile, match *models.ShopMatch, options models.PrintOrderOptions) float64 {
	score := 0.0

	// 1. Price competitiveness (40% weight)
	// Normalize price: lower price = higher score
	// Use inverse relationship: score = 100 / (1 + price/100)
	priceScore := 100.0 / (1.0 + match.TotalPrice/100.0)
	score += priceScore * 0.4

	// 2. Shop rating (30% weight)
	// Normalize rating (assuming 0-5 scale)
	ratingScore := (shop.Rating / 5.0) * 100.0
	score += ratingScore * 0.3

	// 3. Delivery time (20% weight)
	// Faster delivery = higher score
	// Normalize: score = 100 * (1 - deliveryDays/10)
	deliveryScore := math.Max(0, 100.0*(1.0-float64(match.DeliveryDays)/10.0))
	score += deliveryScore * 0.2

	// 4. Service quality / Technology (10% weight)
	// Premium technologies get higher scores
	techScore := m.getTechnologyScore(match.Technology)
	score += techScore * 0.1

	return score
}

// getTechnologyScore returns a score based on technology type
func (m *SmartMatcher) getTechnologyScore(technology string) float64 {
	techScores := map[string]float64{
		"giclée":          90.0,
		"dye-sublimation": 85.0,
		"inkjet":          80.0,
		"laser":           75.0,
		"offset":          70.0,
	}

	if score, ok := techScores[technology]; ok {
		return score
	}
	return 70.0 // Default score
}
