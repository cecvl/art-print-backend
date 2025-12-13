package pricing

import (
	"math"

	"github.com/cecvl/art-print-backend/internal/interfaces"
	"github.com/cecvl/art-print-backend/internal/models"
)

type PricingService struct{}

func NewPricingService() *PricingService {
	return &PricingService{}
}

type PriceRequest struct {
	Size      string `json:"size"`
	Frame     string `json:"frame"`
	Material  string `json:"material"`
	Medium    string `json:"medium"`
	Quantity  int    `json:"quantity"`
	RushOrder bool   `json:"rushOrder"`
}

type PriceResponse struct {
	Total int `json:"total"`
}

// CalculateFramePrice calculates the price for a frame based on size
// Uses size-specific pricing if available, otherwise falls back to base price
func (p *PricingService) CalculateFramePrice(frame *models.Frame, size string) float64 {
	// Check if there's size-specific pricing for this frame
	if frame.SizePricing != nil {
		if price, ok := frame.SizePricing[size]; ok {
			return price
		}
	}
	// Fall back to base price
	return frame.BasePrice
}

// Calculate uses the hardcoded catalog (backward compatibility)
// This will be phased out as shop-specific pricing is adopted
func (p *PricingService) Calculate(req PriceRequest, opts interfaces.PrintOptionsResponse) PriceResponse {
	base := 500.0 // base charge for printing

	// Find size multiplier
	for _, s := range opts.Sizes {
		if s.Name == req.Size {
			base *= s.Multiplier
			break
		}
	}

	// Find frame price
	for _, f := range opts.Frames {
		if f.Type == req.Frame {
			base += float64(f.Price)
			break
		}
	}

	// Find material multiplier
	for _, m := range opts.Materials {
		if m.Type == req.Material {
			base *= m.Multiplier
			break
		}
	}

	// Find medium base
	for _, md := range opts.Mediums {
		if md.Type == req.Medium {
			base += float64(md.BasePrice)
			break
		}
	}

	// Apply quantity if provided
	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}
	base *= float64(quantity)

	return PriceResponse{Total: int(math.Round(base))}
}

// CalculateShopPrice calculates price using a shop's service size-based pricing
// This is the new shop-specific pricing method
func (p *PricingService) CalculateShopPrice(service *models.PrintService, options models.PrintOrderOptions) float64 {
	// Get the total price for the requested size
	price, ok := service.SizePricing[options.Size]
	if !ok {
		return 0 // Size not available for this service
	}

	// Note: Frame pricing is handled separately outside this function
	// Each frame can have its own size-specific pricing

	// Apply quantity discount if configured
	for _, tier := range service.QuantityTiers {
		if options.Quantity >= tier.MinQuantity && (tier.MaxQuantity == 0 || options.Quantity <= tier.MaxQuantity) {
			price *= (1.0 - tier.Discount) // Apply discount
			break
		}
	}

	// Add rush order fee if applicable
	if options.RushOrder {
		price += service.RushOrderFee
	}

	// Multiply by quantity
	price *= float64(options.Quantity)

	return price
}

// CalculateShopPriceWithBreakdown returns price with detailed breakdown
type PriceBreakdown struct {
	BasePrice        float64 `json:"basePrice"`
	SizeModifier     float64 `json:"sizeModifier"`
	MaterialMarkup   float64 `json:"materialMarkup"`
	MediumMarkup     float64 `json:"mediumMarkup"`
	FramePrice       float64 `json:"framePrice"`
	Quantity         int     `json:"quantity"`
	QuantityDiscount float64 `json:"quantityDiscount"`
	RushOrderFee     float64 `json:"rushOrderFee"`
	Subtotal         float64 `json:"subtotal"`
	Total            float64 `json:"total"`
}

func (p *PricingService) CalculateShopPriceWithBreakdown(service *models.PrintService, options models.PrintOrderOptions) PriceBreakdown {
	// Get the size-specific price
	sizePrice, ok := service.SizePricing[options.Size]
	if !ok {
		return PriceBreakdown{} // Size not available
	}

	breakdown := PriceBreakdown{
		BasePrice: sizePrice,
		Quantity:  options.Quantity,
	}

	price := sizePrice

	// Note: Frame price is tracked but added separately
	// Frame pricing is kept separate as frames have their own size-specific pricing
	breakdown.Subtotal = price

	// Apply quantity discount if applicable
	for _, tier := range service.QuantityTiers {
		if options.Quantity >= tier.MinQuantity && (tier.MaxQuantity == 0 || options.Quantity <= tier.MaxQuantity) {
			breakdown.QuantityDiscount = tier.Discount
			price *= (1.0 - tier.Discount)
			break
		}
	}

	// Add rush order fee if applicable
	if options.RushOrder {
		breakdown.RushOrderFee = service.RushOrderFee
		price += service.RushOrderFee
	}

	// Multiply by quantity
	price *= float64(options.Quantity)
	breakdown.Total = price

	return breakdown
}
