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
	json.NewEncoder(w).Encode(service.PriceMatrix)
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

	// Parse price matrix update
	var priceMatrix models.PriceMatrix
	if err := json.NewDecoder(r.Body).Decode(&priceMatrix); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update service with new price matrix
	updates := map[string]interface{}{
		"priceMatrix": priceMatrix,
	}

	if err := h.repo.UpdateService(ctx, serviceID, updates); err != nil {
		log.Printf("❌ Failed to update pricing: %v", err)
		http.Error(w, "Failed to update pricing", http.StatusInternalServerError)
		return
	}

	// Return updated service
	updatedService, _ := h.repo.GetServiceByID(ctx, serviceID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedService.PriceMatrix)
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
