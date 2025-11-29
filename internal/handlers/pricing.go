package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/catalog"
	"github.com/cecvl/art-print-backend/internal/services/pricing"
)

type PricingHandler struct {
	catalog *catalog.CatalogService
	pricing *pricing.PricingService
	repo    *repositories.PrintShopRepository
}

func NewPricingHandler() *PricingHandler {
	return &PricingHandler{
		catalog: catalog.NewCatalogService(),
		pricing: pricing.NewPricingService(),
		repo:    repositories.NewPrintShopRepository(firebase.FirestoreClient),
	}
}

// CalculatePrice calculates price - supports both legacy catalog and shop-specific pricing
// If serviceId is provided, uses shop-specific pricing; otherwise uses legacy catalog
func (h *PricingHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		pricing.PriceRequest
		ServiceID string `json:"serviceId"` // Optional: if provided, use shop-specific pricing
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If serviceId is provided, use shop-specific pricing
	if req.ServiceID != "" {
		ctx := r.Context()
		service, err := h.repo.GetServiceByID(ctx, req.ServiceID)
		if err != nil {
			log.Printf("❌ Service not found: %v", err)
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		options := models.PrintOrderOptions{
			Size:      req.Size,
			Material:  req.Material,
			Medium:    req.Medium,
			Frame:     req.Frame,
			Quantity:  req.Quantity,
			RushOrder: req.RushOrder,
		}

		if options.Quantity == 0 {
			options.Quantity = 1
		}

		totalPrice := h.pricing.CalculateShopPrice(service, options)
		breakdown := h.pricing.CalculateShopPriceWithBreakdown(service, options)

		response := map[string]interface{}{
			"serviceId":  req.ServiceID,
			"total":     int(totalPrice),
			"totalPrice": totalPrice,
			"breakdown": breakdown,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Legacy: use hardcoded catalog (backward compatibility)
	opts := h.catalog.GetPrintOptions()
	result := h.pricing.Calculate(req.PriceRequest, opts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
