package pricing

import (
	"math"

	"github.com/cecvl/art-print-backend/internal/interfaces"
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
}

type PriceResponse struct {
	Total int `json:"total"`
}

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

	return PriceResponse{Total: int(math.Round(base))}
}
