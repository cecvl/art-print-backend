package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/google/uuid"
)

// CheckoutHandler converts the user's cart into an order
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	cartDoc, err := fsClient.Collection("carts").Doc(buyerID).Get(ctx)
	if err != nil || !cartDoc.Exists() {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	var cart models.Cart
	_ = cartDoc.DataTo(&cart)

	var total float64
	for _, item := range cart.Items {
		total += item.Price * float64(item.Quantity)
	}

	order := models.Order{
		OrderID:       uuid.NewString(),
		BuyerID:       buyerID,
		Items:         cart.Items,
		TotalAmount:   total,
		Status:        "pending",
		PaymentMethod: "unpaid",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// ✅ Correct assignment for Firestore Set
	if _, err := fsClient.Collection("orders").Doc(order.OrderID).Set(ctx, order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	// ✅ Delete returns only an error
	if _, err := fsClient.Collection("carts").Doc(buyerID).Delete(ctx); err != nil {
		http.Error(w, "failed to clear cart", http.StatusInternalServerError)
		return
	}

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
