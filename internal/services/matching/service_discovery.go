package matching

import (
	"context"
	"log"

	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/pricing"
)

// ServiceDiscovery helps find shops and services that match order requirements
type ServiceDiscovery struct {
	repo    *repositories.PrintShopRepository
	pricing *pricing.PricingService
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(repo *repositories.PrintShopRepository) *ServiceDiscovery {
	return &ServiceDiscovery{
		repo:    repo,
		pricing: pricing.NewPricingService(),
	}
}

// FindMatchingShops finds all shops that can fulfill an order with given options
func (sd *ServiceDiscovery) FindMatchingShops(ctx context.Context, options models.PrintOrderOptions) ([]models.ShopMatch, error) {
	// Get all active shops
	shops, err := sd.repo.GetActiveShops(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]models.ShopMatch, 0)

	// For each shop, find matching services
	for _, shop := range shops {
		services, err := sd.repo.GetServicesByShopID(ctx, shop.ID)
		if err != nil {
			log.Printf("⚠️ Failed to get services for shop %s: %v", shop.ID, err)
			continue
		}

		// Check each service for compatibility
		for _, service := range services {
			if !service.IsActive {
				continue
			}

			// Check if service supports the requested options
			if sd.serviceSupportsOptions(service, options) {
				totalPrice := sd.pricing.CalculateShopPrice(service, options)

				// Calculate match score (basic - will be enhanced in smart matcher)
				matchScore := sd.calculateBasicScore(shop, service, totalPrice)

				matches = append(matches, models.ShopMatch{
					ShopID:       shop.ID,
					ShopName:     shop.Name,
					ServiceID:    service.ID,
					TotalPrice:   totalPrice,
					MatchScore:   matchScore,
					Technology:   service.Technology,
					DeliveryDays: sd.estimateDeliveryDays(shop, service, options),
				})
			}
		}
	}

	return matches, nil
}

// serviceSupportsOptions checks if a service can fulfill the order options
func (sd *ServiceDiscovery) serviceSupportsOptions(service *models.PrintService, options models.PrintOrderOptions) bool {
	matrix := service.PriceMatrix

	// Check size support (if sizeModifiers is empty, supports all sizes)
	if len(matrix.SizeModifiers) > 0 {
		if _, ok := matrix.SizeModifiers[options.Size]; !ok {
			return false
		}
	}

	// Check material support
	if len(matrix.MaterialMarkups) > 0 {
		if _, ok := matrix.MaterialMarkups[options.Material]; !ok {
			return false
		}
	}

	// Check medium support
	if len(matrix.MediumMarkups) > 0 {
		if _, ok := matrix.MediumMarkups[options.Medium]; !ok {
			return false
		}
	}

	// Check frame support
	if len(matrix.FramePrices) > 0 {
		if _, ok := matrix.FramePrices[options.Frame]; !ok {
			return false
		}
	}

	return true
}

// calculateBasicScore calculates a basic match score (used by auto matcher)
func (sd *ServiceDiscovery) calculateBasicScore(shop *models.PrintShopProfile, service *models.PrintService, price float64) float64 {
	// Simple score: lower price = higher score
	// Normalize to 0-100 range
	if price <= 0 {
		return 0
	}
	return 100.0 / (1.0 + price/100.0)
}

// estimateDeliveryDays estimates delivery time (can be enhanced with real data)
func (sd *ServiceDiscovery) estimateDeliveryDays(shop *models.PrintShopProfile, service *models.PrintService, options models.PrintOrderOptions) int {
	baseDays := 5 // Base delivery time

	// Rush orders are faster
	if options.RushOrder {
		baseDays = 2
	}

	// Larger quantities take longer
	if options.Quantity > 20 {
		baseDays += 2
	}

	return baseDays
}
