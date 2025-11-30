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

// PrintShopConsoleHandler handles print shop console operations
type PrintShopConsoleHandler struct {
	repo *repositories.PrintShopRepository
}

// NewPrintShopConsoleHandler creates a new print shop console handler
func NewPrintShopConsoleHandler() *PrintShopConsoleHandler {
	return &PrintShopConsoleHandler{
		repo: repositories.NewPrintShopRepository(firebase.FirestoreClient),
	}
}

// GetShopProfile retrieves the print shop profile for the authenticated owner
func (h *PrintShopConsoleHandler) GetShopProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		log.Printf("❌ Failed to get shop: %v", err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

// UpdateShopProfile updates the print shop profile
func (h *PrintShopConsoleHandler) UpdateShopProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Get existing shop
	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		log.Printf("❌ Shop not found: %v", err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Parse update request
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Don't allow updating ownerId or ID
	delete(updates, "ownerId")
	delete(updates, "id")

	// Update shop
	if err := h.repo.UpdateShop(ctx, shop.ID, updates); err != nil {
		log.Printf("❌ Failed to update shop: %v", err)
		http.Error(w, "Failed to update shop", http.StatusInternalServerError)
		return
	}

	// Return updated shop
	updatedShop, _ := h.repo.GetShopByID(ctx, shop.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedShop)
}

// CreateShopProfile creates a new print shop profile (for initial setup)
func (h *PrintShopConsoleHandler) CreateShopProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Check if shop already exists
	_, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err == nil {
		http.Error(w, "Shop profile already exists", http.StatusConflict)
		return
	}

	var shop models.PrintShopProfile
	if err := json.NewDecoder(r.Body).Decode(&shop); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set owner ID and ensure it matches authenticated user
	shop.OwnerID = ownerID
	shop.IsActive = true
	shop.Services = []string{} // Initialize empty services list

	// Create shop
	if err := h.repo.CreateShop(ctx, &shop); err != nil {
		log.Printf("❌ Failed to create shop: %v", err)
		http.Error(w, "Failed to create shop", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shop)
}

// ==================== Service Management ====================

// GetServices retrieves all services for the authenticated shop
func (h *PrintShopConsoleHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Get shop first
	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		log.Printf("❌ Shop not found: %v", err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	// Get all services
	services, err := h.repo.GetServicesByShopID(ctx, shop.ID)
	if err != nil {
		log.Printf("❌ Failed to get services: %v", err)
		http.Error(w, "Failed to get services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// CreateService creates a new print service
func (h *PrintShopConsoleHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Get shop first
	shop, err := h.repo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		log.Printf("❌ Shop not found: %v", err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	var service models.PrintService
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set shop ID and ensure it matches authenticated shop
	service.ShopID = shop.ID
	service.IsActive = true

	// Create service
	if err := h.repo.CreateService(ctx, &service); err != nil {
		log.Printf("❌ Failed to create service: %v", err)
		http.Error(w, "Failed to create service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(service)
}

// UpdateService updates a print service
func (h *PrintShopConsoleHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Extract service ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	serviceID := pathParts[len(pathParts)-1]

	// Get service to verify ownership
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		log.Printf("❌ Service not found: %v", err)
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Verify shop ownership
	shop, err := h.repo.GetShopByID(ctx, service.ShopID)
	if err != nil || shop.OwnerID != ownerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse updates
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Don't allow updating shopId or ID
	delete(updates, "shopId")
	delete(updates, "id")

	// Update service
	if err := h.repo.UpdateService(ctx, serviceID, updates); err != nil {
		log.Printf("❌ Failed to update service: %v", err)
		http.Error(w, "Failed to update service", http.StatusInternalServerError)
		return
	}

	// Return updated service
	updatedService, _ := h.repo.GetServiceByID(ctx, serviceID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedService)
}

// DeleteService deletes a print service
func (h *PrintShopConsoleHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID := ctx.Value("shopOwnerId").(string)

	// Extract service ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	serviceID := pathParts[len(pathParts)-1]

	// Get service to verify ownership
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		log.Printf("❌ Service not found: %v", err)
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Verify shop ownership
	shop, err := h.repo.GetShopByID(ctx, service.ShopID)
	if err != nil || shop.OwnerID != ownerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete service
	if err := h.repo.DeleteService(ctx, serviceID); err != nil {
		log.Printf("❌ Failed to delete service: %v", err)
		http.Error(w, "Failed to delete service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
