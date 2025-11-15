package orders

import (
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

func (s *OrderService) AssignShopForOrder(order *models.Order) error {
	mode := s.configService.GetFulfillmentMode()

	switch mode {
	case models.FulfillmentAuto:
		return s.autoMatcher.Assign(order)
	case models.FulfillmentManual:
		return s.manualMatcher.Assign(order)
	case models.FulfillmentSmart:
		return s.smartMatcher.Assign(order)
	default:
		return s.autoMatcher.Assign(order)
	}
}
