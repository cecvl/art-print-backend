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
	Size     string `json:"size"`
	Frame    string `json:"frame"`
	Material string `json:"material"`
	Medium   string `json:"medium"`
	Quantity int    `json:"quantity"`
	RushOrder bool  `json:"rushOrder"`
}

type PriceResponse struct {
	Total int `json:"total"`
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

// CalculateShopPrice calculates price using a shop's service PriceMatrix
// This is the new shop-specific pricing method using Go's computational efficiency
func (p *PricingService) CalculateShopPrice(service *models.PrintService, options models.PrintOrderOptions) float64 {
	matrix := service.PriceMatrix
	price := service.BasePrice

	// Apply size modifier
	if modifier, ok := matrix.SizeModifiers[options.Size]; ok {
		price *= modifier
	}

	// Apply material markup
	if markup, ok := matrix.MaterialMarkups[options.Material]; ok {
		price *= markup
	}

	// Apply medium markup
	if markup, ok := matrix.MediumMarkups[options.Medium]; ok {
		price *= markup
	}

	// Add frame price
	if framePrice, ok := matrix.FramePrices[options.Frame]; ok {
		price += framePrice
	}

	// Apply quantity discount
	for _, tier := range matrix.QuantityTiers {
		if options.Quantity >= tier.MinQuantity && options.Quantity <= tier.MaxQuantity {
			price *= (1.0 - tier.Discount) // Apply discount
			break
		}
	}

	// Add rush order fee
	if options.RushOrder {
		price += matrix.RushOrderFee
	}

	// Multiply by quantity
	price *= float64(options.Quantity)

	return price
}

// CalculateShopPriceWithBreakdown returns price with detailed breakdown
type PriceBreakdown struct {
	BasePrice      float64            `json:"basePrice"`
	SizeModifier   float64            `json:"sizeModifier"`
	MaterialMarkup float64            `json:"materialMarkup"`
	MediumMarkup   float64            `json:"mediumMarkup"`
	FramePrice     float64            `json:"framePrice"`
	Quantity       int                `json:"quantity"`
	QuantityDiscount float64          `json:"quantityDiscount"`
	RushOrderFee   float64            `json:"rushOrderFee"`
	Subtotal       float64            `json:"subtotal"`
	Total          float64            `json:"total"`
}

func (p *PricingService) CalculateShopPriceWithBreakdown(service *models.PrintService, options models.PrintOrderOptions) PriceBreakdown {
	matrix := service.PriceMatrix
	breakdown := PriceBreakdown{
		BasePrice: service.BasePrice,
		Quantity:  options.Quantity,
	}

	price := service.BasePrice

	// Apply size modifier
	if modifier, ok := matrix.SizeModifiers[options.Size]; ok {
		breakdown.SizeModifier = modifier
		price *= modifier
	}

	// Apply material markup
	if markup, ok := matrix.MaterialMarkups[options.Material]; ok {
		breakdown.MaterialMarkup = markup
		price *= markup
	}

	// Apply medium markup
	if markup, ok := matrix.MediumMarkups[options.Medium]; ok {
		breakdown.MediumMarkup = markup
		price *= markup
	}

	// Add frame price
	if framePrice, ok := matrix.FramePrices[options.Frame]; ok {
		breakdown.FramePrice = framePrice
		price += framePrice
	}

	breakdown.Subtotal = price

	// Apply quantity discount
	for _, tier := range matrix.QuantityTiers {
		if options.Quantity >= tier.MinQuantity && options.Quantity <= tier.MaxQuantity {
			breakdown.QuantityDiscount = tier.Discount
			price *= (1.0 - tier.Discount)
			break
		}
	}

	// Add rush order fee
	if options.RushOrder {
		breakdown.RushOrderFee = matrix.RushOrderFee
		price += matrix.RushOrderFee
	}

	// Multiply by quantity
	price *= float64(options.Quantity)
	breakdown.Total = price

	return breakdown
}
