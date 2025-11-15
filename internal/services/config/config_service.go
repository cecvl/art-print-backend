package config

import "github.com/cecvl/art-print-backend/internal/models"

type ConfigService interface {
	GetFulfillmentMode() models.FulfillmentMode
}

type DefaultConfigService struct {
	Mode models.FulfillmentMode
}

func NewDefaultConfigService() *DefaultConfigService {
	return &DefaultConfigService{
		Mode: models.FulfillmentAuto, // fallback default
	}
}

func (c *DefaultConfigService) GetFulfillmentMode() models.FulfillmentMode {
	// later: load from Firestore (settings/global)
	return c.Mode
}

// For admin to toggle later
func (c *DefaultConfigService) SetFulfillmentMode(mode models.FulfillmentMode) {
	c.Mode = mode
}
