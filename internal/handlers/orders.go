package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/services/config"
	"github.com/cecvl/art-print-backend/internal/services/orders"
	"github.com/google/uuid"
)

// CheckoutHandler converts the user's cart into an order and assigns a print shop
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	// Parse request to get print options
	var checkoutReq struct {
		PrintOptions models.PrintOrderOptions `json:"printOptions"`
	}

	// Try to decode print options (optional - can use defaults)
	json.NewDecoder(r.Body).Decode(&checkoutReq)

	cartDoc, err := fsClient.Collection("carts").Doc(buyerID).Get(ctx)
	if err != nil || !cartDoc.Exists() {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	var cart models.Cart
	_ = cartDoc.DataTo(&cart)

	// If print options not provided, try to extract from first cart item
	if checkoutReq.PrintOptions.Size == "" && len(cart.Items) > 0 && cart.Items[0].PrintOptions.Size != "" {
		checkoutReq.PrintOptions = cart.Items[0].PrintOptions
	}

	// Set defaults if still empty
	if checkoutReq.PrintOptions.Quantity == 0 {
		checkoutReq.PrintOptions.Quantity = 1
		if len(cart.Items) > 0 {
			checkoutReq.PrintOptions.Quantity = cart.Items[0].Quantity
		}
	}

	var total float64
	for _, item := range cart.Items {
		total += item.Price * float64(item.Quantity)
	}

	order := models.Order{
		OrderID:        uuid.NewString(),
		BuyerID:        buyerID,
		Items:          cart.Items,
		PrintOptions:   checkoutReq.PrintOptions,
		TotalAmount:    total,
		Status:         "pending",
		PaymentMethod:  "unpaid",  // Legacy field
		PaymentStatus:  "unpaid",  // New field
		DeliveryStatus: "pending", // Initialize delivery status
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Assign print shop using matching service
	configService := config.NewDefaultConfigService()
	orderService := orders.NewOrderService(configService)

	if err := orderService.AssignShopForOrder(ctx, &order); err != nil {
		log.Printf("⚠️ Failed to assign shop for order %s: %v", order.OrderID, err)
		// Continue without assignment - order can be manually assigned later
	}

	// Save order to Firestore
	if _, err := fsClient.Collection("orders").Doc(order.OrderID).Set(ctx, order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	// Clear cart
	if _, err := fsClient.Collection("carts").Doc(buyerID).Delete(ctx); err != nil {
		log.Printf("⚠️ Failed to clear cart: %v", err)
		// Don't fail the request if cart clearing fails
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetOrdersHandler fetches all orders for the authenticated user
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	snaps, err := fsClient.Collection("orders").Where("buyerId", "==", buyerID).Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, "failed to fetch orders", http.StatusInternalServerError)
		return
	}

	var orders []models.Order
	for _, snap := range snaps {
		var o models.Order
		_ = snap.DataTo(&o)
		orders = append(orders, o)
	}

	json.NewEncoder(w).Encode(orders)
}
