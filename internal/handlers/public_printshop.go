package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/pricing"
)

// PublicPrintShopHandler handles public endpoints for print shop discovery
type PublicPrintShopHandler struct {
	repo    *repositories.PrintShopRepository
	pricing *pricing.PricingService
}

// NewPublicPrintShopHandler creates a new public print shop handler
func NewPublicPrintShopHandler() *PublicPrintShopHandler {
	return &PublicPrintShopHandler{
		repo:    repositories.NewPrintShopRepository(firebase.FirestoreClient),
		pricing: pricing.NewPricingService(),
	}
}

// GetActiveShops returns all active print shops (public endpoint)
func (h *PublicPrintShopHandler) GetActiveShops(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shops, err := h.repo.GetActiveShops(ctx)
	if err != nil {
		log.Printf("❌ Failed to get active shops: %v", err)
		http.Error(w, "Failed to get shops", http.StatusInternalServerError)
		return
	}

	// Return simplified shop info (no sensitive data)
	type ShopSummary struct {
		ID           string          `json:"id"`
		Name         string          `json:"name"`
		Description  string          `json:"description"`
		Location     models.Location `json:"location"`
		Rating       float64         `json:"rating"`
		ServiceCount int             `json:"serviceCount"`
	}

	summaries := make([]ShopSummary, 0, len(shops))
	for _, shop := range shops {
		summaries = append(summaries, ShopSummary{
			ID:           shop.ID,
			Name:         shop.Name,
			Description:  shop.Description,
			Location:     shop.Location,
			Rating:       shop.Rating,
			ServiceCount: len(shop.Services),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// GetShopDetails returns detailed shop information including services
func (h *PublicPrintShopHandler) GetShopDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract shop ID from URL
	shopID := r.URL.Query().Get("id")
	if shopID == "" {
		http.Error(w, "Shop ID required", http.StatusBadRequest)
		return
	}

	shop, err := h.repo.GetShopByID(ctx, shopID)
	if err != nil {
		log.Printf("❌ Shop not found: %v", err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	if !shop.IsActive {
		http.Error(w, "Shop not available", http.StatusNotFound)
		return
	}

	// Get shop services
	services, err := h.repo.GetServicesByShopID(ctx, shopID)
	if err != nil {
		log.Printf("⚠️ Failed to get services: %v", err)
		services = []*models.PrintService{} // Return empty if error
	}

	// Return public shop info
	type ShopDetails struct {
		ID          string                 `json:"id"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Location    models.Location        `json:"location"`
		Contact     models.ContactInfo     `json:"contact"`
		Rating      float64                `json:"rating"`
		Services    []*models.PrintService `json:"services"`
	}

	details := ShopDetails{
		ID:          shop.ID,
		Name:        shop.Name,
		Description: shop.Description,
		Location:    shop.Location,
		Contact:     shop.Contact,
		Rating:      shop.Rating,
		Services:    services,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// CalculatePriceForService calculates price for a specific service (public endpoint)
func (h *PublicPrintShopHandler) CalculatePriceForService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ServiceID string                   `json:"serviceId"`
		Options   models.PrintOrderOptions `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get service
	service, err := h.repo.GetServiceByID(ctx, req.ServiceID)
	if err != nil {
		log.Printf("❌ Service not found: %v", err)
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Verify service is active
	if !service.IsActive {
		http.Error(w, "Service not available", http.StatusNotFound)
		return
	}

	// Calculate price
	totalPrice := h.pricing.CalculateShopPrice(service, req.Options)
	breakdown := h.pricing.CalculateShopPriceWithBreakdown(service, req.Options)

	// Get shop info
	shop, _ := h.repo.GetShopByID(ctx, service.ShopID)

	response := map[string]interface{}{
		"serviceId":   req.ServiceID,
		"shopId":      service.ShopID,
		"shopName":    shop.Name,
		"serviceName": service.Name,
		"options":     req.Options,
		"breakdown":   breakdown,
		"totalPrice":  totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MatchShopsForOrder finds shops that can fulfill an order (public endpoint)
func (h *PublicPrintShopHandler) MatchShopsForOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var options models.PrintOrderOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get all active shops
	shops, err := h.repo.GetActiveShops(ctx)
	if err != nil {
		log.Printf("❌ Failed to get shops: %v", err)
		http.Error(w, "Failed to get shops", http.StatusInternalServerError)
		return
	}

	matches := make([]models.ShopMatch, 0)

	// For each shop, find matching services
	for _, shop := range shops {
		services, err := h.repo.GetServicesByShopID(ctx, shop.ID)
		if err != nil {
			continue
		}

		// Check each service for compatibility
		for _, service := range services {
			if !service.IsActive {
				continue
			}

			// Check if service supports the requested options
			// matrix := service.PriceMatrix // Deprecated
			
			// Check size availability
			_, hasSize := service.SizePricing[options.Size]
			
			// TODO: Resolve SubstrateID and MediumID to check against options.Material and options.Medium
			// For now, we assume if the service is active and has the size, it's a candidate
			// This needs to be updated to fetch Material/Medium details or cache them
			hasMaterial := true 
			hasMedium := true
			hasFrame := true // Frame check also needs update as frames are separate

			// If service supports all options, calculate price
			if hasSize && hasMaterial && hasMedium && hasFrame {
				totalPrice := h.pricing.CalculateShopPrice(service, options)

				// Calculate match score (simple: based on price competitiveness)
				// Lower price = higher score (will be enhanced in smart matching)
				matchScore := 100.0 / (1.0 + totalPrice/100.0) // Normalize score

				techType := ""
				if service.Technology != nil {
					techType = service.Technology.Type
				}

				matches = append(matches, models.ShopMatch{
					ShopID:       shop.ID,
					ShopName:     shop.Name,
					ServiceID:    service.ID,
					TotalPrice:   totalPrice,
					MatchScore:   matchScore,
					Technology:   techType,
					DeliveryDays: 5, // Default, can be enhanced later
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
