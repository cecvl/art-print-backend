package orders

import (
	"context"

	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/services/config"
	"github.com/cecvl/art-print-backend/internal/services/matching"
)

type OrderService struct {
	configService config.ConfigService

	autoMatcher   *matching.AutoMatcher
	manualMatcher *matching.ManualMatcher
	smartMatcher  *matching.SmartMatcher
}

func NewOrderService(cfg config.ConfigService) *OrderService {
	return &OrderService{
		configService: cfg,
		autoMatcher:   matching.NewAutoMatcher(),
		manualMatcher: matching.NewManualMatcher(),
		smartMatcher:  matching.NewSmartMatcher(),
	}
}

// AssignShopForOrder assigns a print shop to an order based on fulfillment mode
func (s *OrderService) AssignShopForOrder(ctx context.Context, order *models.Order) error {
	mode := s.configService.GetFulfillmentMode()

	// Extract print options from order
	options := order.PrintOptions

	switch mode {
	case models.FulfillmentAuto:
		return s.autoMatcher.Assign(ctx, order, options)
	case models.FulfillmentManual:
		return s.manualMatcher.Assign(ctx, order, options)
	case models.FulfillmentSmart:
		return s.smartMatcher.Assign(ctx, order, options)
	default:
		return s.autoMatcher.Assign(ctx, order, options)
	}
}

// GetMatchesForOrder returns all matching shops for an order (useful for manual mode)
func (s *OrderService) GetMatchesForOrder(ctx context.Context, order *models.Order) ([]models.ShopMatch, error) {
	options := order.PrintOptions
	return s.manualMatcher.GetMatches(ctx, order, options)
}
