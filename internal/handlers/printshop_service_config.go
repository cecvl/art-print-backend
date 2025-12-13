package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// PrintShopServiceConfigHandler handles service pricing configuration
type PrintShopServiceConfigHandler struct {
	repo *repositories.PrintShopRepository
}

// NewPrintShopServiceConfigHandler creates a new service config handler
func NewPrintShopServiceConfigHandler() *PrintShopServiceConfigHandler {
	return &PrintShopServiceConfigHandler{
		repo: repositories.NewPrintShopRepository(firebase.FirestoreClient),
	}
}

// GetServicePricing retrieves the price matrix for a service
func (h *PrintShopServiceConfigHandler) GetServicePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	serviceID := pathParts[len(pathParts)-2] // /api/printshop/services/{id}/pricing

	// Get service and verify ownership
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	shop, err := h.repo.GetShopByID(ctx, service.ShopID)
	if err != nil || shop.OwnerID != ownerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sizePricing":   service.SizePricing,
		"quantityTiers": service.QuantityTiers,
		"rushOrderFee":  service.RushOrderFee,
	})
}

// UpdateServicePricing updates the price matrix for a service
func (h *PrintShopServiceConfigHandler) UpdateServicePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	serviceID := pathParts[len(pathParts)-2] // /api/printshop/services/{id}/pricing

	// Get service and verify ownership
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	shop, err := h.repo.GetShopByID(ctx, service.ShopID)
	if err != nil || shop.OwnerID != ownerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse pricing update
	var payload struct {
		SizePricing   map[string]float64    `json:"sizePricing"`
		QuantityTiers []models.QuantityTier `json:"quantityTiers,omitempty"`
		RushOrderFee  float64               `json:"rushOrderFee,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate size pricing is provided
	if len(payload.SizePricing) == 0 {
		http.Error(w, "sizePricing is required", http.StatusBadRequest)
		return
	}

	// Update service with new pricing
	updates := map[string]interface{}{
		"sizePricing": payload.SizePricing,
	}
	if payload.QuantityTiers != nil {
		updates["quantityTiers"] = payload.QuantityTiers
	}
	if payload.RushOrderFee > 0 {
		updates["rushOrderFee"] = payload.RushOrderFee
	}

	if err := h.repo.UpdateService(ctx, serviceID, updates); err != nil {
		log.Printf("❌ Failed to update pricing: %v", err)
		http.Error(w, "Failed to update pricing", http.StatusInternalServerError)
		return
	}

	// Return updated service
	updatedService, _ := h.repo.GetServiceByID(ctx, serviceID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sizePricing":   updatedService.SizePricing,
		"quantityTiers": updatedService.QuantityTiers,
		"rushOrderFee":  updatedService.RushOrderFee,
	})
}

// CalculateServicePrice calculates price for given options (for testing/validation)
func (h *PrintShopServiceConfigHandler) CalculateServicePrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	serviceID := pathParts[len(pathParts)-2] // /api/printshop/services/{id}/calculate

	// Get service and verify ownership
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	shop, err := h.repo.GetShopByID(ctx, service.ShopID)
	if err != nil || shop.OwnerID != ownerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse order options
	var options models.PrintOrderOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate price using service's price matrix
	price := calculatePrice(service, options)

	response := map[string]interface{}{
		"serviceId":  serviceID,
		"options":    options,
		"totalPrice": price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculatePrice calculates the total price based on service and options
func calculatePrice(service *models.PrintService, options models.PrintOrderOptions) float64 {
	// Get base price for the size from the service's size pricing
	price, ok := service.SizePricing[options.Size]
	if !ok {
		return 0 // Size not supported by this service
	}

	// Note: Frame pricing is handled separately and added by the caller if needed
	// This allows frames to have their own size-specific pricing

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
