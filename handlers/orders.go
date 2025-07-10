package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/models"
)

// === Cart ===

func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	var item models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid cart item", http.StatusBadRequest)
		return
	}

	itemRef := firebase.FirestoreClient.Collection("users").Doc(userID).Collection("cartItems").Doc(item.ArtworkID)
	_, err := itemRef.Set(ctx, item)
	if err != nil {
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	docs, err := firebase.FirestoreClient.Collection("users").Doc(userID).Collection("cartItems").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve cart", http.StatusInternalServerError)
		return
	}

	var cart []models.CartItem
	for _, doc := range docs {
		var item models.CartItem
		if err := doc.DataTo(&item); err == nil {
			cart = append(cart, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	artworkID := r.URL.Query().Get("artworkId")
	if artworkID == "" {
		http.Error(w, "Missing artworkId", http.StatusBadRequest)
		return
	}

	_, err := firebase.FirestoreClient.Collection("users").Doc(userID).Collection("cartItems").Doc(artworkID).Delete(ctx)
	if err != nil {
		http.Error(w, "Failed to remove item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// === Checkout ===

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	cartDocs, err := firebase.FirestoreClient.Collection("users").Doc(userID).Collection("cartItems").Documents(ctx).GetAll()
	if err != nil || len(cartDocs) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	var items []models.CartItem
	var total float64
	for _, doc := range cartDocs {
		var item models.CartItem
		if err := doc.DataTo(&item); err == nil {
			items = append(items, item)
			total += item.Price * float64(item.Quantity)
		}
	}

	printShops, err := firebase.FirestoreClient.Collection("users").
		Where("roles", "array-contains", models.PrintShop).
		Documents(ctx).GetAll()
	if err != nil || len(printShops) == 0 {
		http.Error(w, "No print shops available", http.StatusServiceUnavailable)
		return
	}

	order := models.Order{
		BuyerID:       userID,
		PrintShopID:   printShops[0].Ref.ID,
		Items:         items,
		TotalAmount:   total,
		PaymentMethod: "manual",
		TransactionID: "TXN-" + time.Now().Format("20060102150405"),
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, _, err = firebase.FirestoreClient.Collection("orders").Add(ctx, order)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Clear the cart
	batch := firebase.FirestoreClient.Batch()
	for _, doc := range cartDocs {
		batch.Delete(doc.Ref)
	}
	_, err = batch.Commit(ctx)
	if err != nil {
		http.Error(w, "Order created but failed to clear cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Order placed",
		"totalAmount": total,
	})
}

// === Legacy Order Creation ===

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	userDoc, err := firebase.FirestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	roles, ok := userDoc.Data()["roles"].([]interface{})
	if !ok || !hasRole(roles, models.Buyer) {
		http.Error(w, "Only buyers can create orders", http.StatusForbidden)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return
	}

	printShops, err := firebase.FirestoreClient.Collection("users").
		Where("roles", "array-contains", models.PrintShop).
		Documents(ctx).GetAll()
	if err != nil || len(printShops) == 0 {
		http.Error(w, "No print shops available", http.StatusServiceUnavailable)
		return
	}

	order.BuyerID = userID
	order.PrintShopID = printShops[0].Ref.ID
	order.Status = "pending"
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, _, err = firebase.FirestoreClient.Collection("orders").Add(ctx, order)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func hasRole(roles []interface{}, role string) bool {
	for _, r := range roles {
		if str, ok := r.(string); ok && str == role {
			return true
		}
	}
	return false
}
