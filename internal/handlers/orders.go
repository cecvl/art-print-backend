package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"google.golang.org/api/iterator"
)

// ===== Checkout order =====
func CheckoutOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	// Find the cart order
	iter := firebase.FirestoreClient.Collection("orders").
		Where("buyerId", "==", userID).
		Where("status", "==", "cart").
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "No cart to checkout", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Error retrieving cart", http.StatusInternalServerError)
		return
	}

	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		http.Error(w, "Error parsing order", http.StatusInternalServerError)
		return
	}

	if len(order.Items) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	// TODO: Validate each artwork again before checkout
	// TODO: Assign print shop dynamically
	order.PrintShopID = "default-printshop-id" // Placeholder

	order.Status = "pending"
	order.UpdatedAt = time.Now()

	_, err = doc.Ref.Set(ctx, order)
	if err != nil {
		http.Error(w, "Failed to checkout", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Order placed",
		"orderId":     doc.Ref.ID,
		"totalAmount": order.TotalAmount,
	})
}

// ===== Get all orders for buyer =====
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	iter := firebase.FirestoreClient.Collection("orders").
		Where("buyerId", "==", userID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var orders []models.Order
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Error retrieving orders", http.StatusInternalServerError)
			return
		}
		var order models.Order
		if err := doc.DataTo(&order); err == nil {
			orders = append(orders, order)
		}
	}

	json.NewEncoder(w).Encode(orders)
}
