package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/config"
	"github.com/cecvl/art-print-backend/internal/services/matching"
	"github.com/cecvl/art-print-backend/internal/services/orders"
)

// MatchingHandler handles order matching operations
type MatchingHandler struct {
	orderService *orders.OrderService
	discovery    *matching.ServiceDiscovery
}

// NewMatchingHandler creates a new matching handler
func NewMatchingHandler() *MatchingHandler {
	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	configService := config.NewDefaultConfigService()
	return &MatchingHandler{
		orderService: orders.NewOrderService(configService),
		discovery:    matching.NewServiceDiscovery(repo),
	}
}

// GetOrderMatches returns all matching shops for an order (for manual selection)
func (h *MatchingHandler) GetOrderMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		OrderID string                   `json:"orderId"`
		Options models.PrintOrderOptions `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a temporary order for matching
	order := &models.Order{
		OrderID:      req.OrderID,
		PrintOptions: req.Options,
	}

	matches, err := h.orderService.GetMatchesForOrder(ctx, order)
	if err != nil {
		log.Printf("❌ Failed to get matches: %v", err)
		http.Error(w, "Failed to get matches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// AssignShopToOrder manually assigns a shop to an order (for manual mode)
func (h *MatchingHandler) AssignShopToOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fsClient := firebase.FirestoreClient

	var req struct {
		OrderID string `json:"orderId"`
		ShopID  string `json:"shopId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get order
	orderDoc, err := fsClient.Collection("orders").Doc(req.OrderID).Get(ctx)
	if err != nil || !orderDoc.Exists() {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	var order models.Order
	if err := orderDoc.DataTo(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusInternalServerError)
		return
	}

	// Verify shop exists and is active
	repo := repositories.NewPrintShopRepository(fsClient)
	shop, err := repo.GetShopByID(ctx, req.ShopID)
	if err != nil || !shop.IsActive {
		http.Error(w, "Shop not found or inactive", http.StatusBadRequest)
		return
	}

	// Update order with shop assignment
	order.PrintShopID = req.ShopID
	order.UpdatedAt = time.Now()

	if _, err := fsClient.Collection("orders").Doc(req.OrderID).Set(ctx, order); err != nil {
		log.Printf("❌ Failed to update order: %v", err)
		http.Error(w, "Failed to assign shop", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
