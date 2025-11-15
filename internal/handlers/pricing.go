package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cecvl/art-print-backend/internal/services/catalog"
	"github.com/cecvl/art-print-backend/internal/services/pricing"
)

type PricingHandler struct {
	catalog *catalog.CatalogService
	pricing *pricing.PricingService
}

func NewPricingHandler() *PricingHandler {
	return &PricingHandler{
		catalog: catalog.NewCatalogService(),
		pricing: pricing.NewPricingService(),
	}
}

func (h *PricingHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	var req pricing.PriceRequest
	json.NewDecoder(r.Body).Decode(&req)

	opts := h.catalog.GetPrintOptions()
	result := h.pricing.Calculate(req, opts)

	json.NewEncoder(w).Encode(result)
}
