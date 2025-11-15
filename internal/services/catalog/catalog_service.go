package catalog

import "github.com/cecvl/art-print-backend/internal/interfaces"

type CatalogService struct{}

func NewCatalogService() *CatalogService {
	return &CatalogService{}
}

func (c *CatalogService) GetPrintOptions() interfaces.PrintOptionsResponse {
	return interfaces.PrintOptionsResponse{
		Sizes: []interfaces.PrintSize{
			{Name: "A4", Multiplier: 1.0, WidthCM: 21, HeightCM: 29},
			{Name: "A3", Multiplier: 1.6, WidthCM: 29, HeightCM: 42},
		},
		Frames: []interfaces.FrameOption{
			{Type: "classic", Material: "wood", Price: 300},
			{Type: "modern", Material: "metal", Price: 450},
		},
		Materials: []interfaces.MaterialOption{
			{Type: "matte", Multiplier: 1.2},
			{Type: "glossy", Multiplier: 1.35},
		},
		Mediums: []interfaces.MediumOption{
			{Type: "canvas", BasePrice: 750},
			{Type: "premium", BasePrice: 1200},
		},
	}
}
