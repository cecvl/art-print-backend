package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// AddToCartHandler adds or updates an item in the user's cart
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	var newItem models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cartRef := fsClient.Collection("carts").Doc(buyerID)
	doc, err := cartRef.Get(ctx)

	var cart models.Cart
	if err == nil && doc.Exists() {
		_ = doc.DataTo(&cart)
		updated := false
		for i, item := range cart.Items {
			if item.ArtworkID == newItem.ArtworkID {
				cart.Items[i].Quantity += newItem.Quantity
				updated = true
				break
			}
		}
		if !updated {
			cart.Items = append(cart.Items, newItem)
		}
	} else {
		cart = models.Cart{
			BuyerID: buyerID,
			Items:   []models.CartItem{newItem},
		}
	}

	cart.UpdatedAt = time.Now()
	if _, err := cartRef.Set(ctx, cart); err != nil {
		http.Error(w, "failed to update cart", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cart)
}

// GetCartHandler retrieves the user's cart
func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	doc, err := fsClient.Collection("carts").Doc(buyerID).Get(ctx)
	if err != nil || !doc.Exists() {
		json.NewEncoder(w).Encode(models.Cart{BuyerID: buyerID, Items: []models.CartItem{}})
		return
	}

	var cart models.Cart
	_ = doc.DataTo(&cart)
	json.NewEncoder(w).Encode(cart)
}

// RemoveFromCartHandler removes an item from the user's cart
func RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fsClient := firebase.GetFirestoreClient()
	buyerID := r.Context().Value("userID").(string)

	var payload struct {
		ArtworkID string `json:"artworkId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cartRef := fsClient.Collection("carts").Doc(buyerID)
	doc, err := cartRef.Get(ctx)
	if err != nil || !doc.Exists() {
		http.Error(w, "cart not found", http.StatusNotFound)
		return
	}

	var cart models.Cart
	_ = doc.DataTo(&cart)

	newItems := []models.CartItem{}
	for _, item := range cart.Items {
		if item.ArtworkID != payload.ArtworkID {
			newItems = append(newItems, item)
		}
	}
	cart.Items = newItems
	cart.UpdatedAt = time.Now()

	if _, err := cartRef.Set(ctx, cart); err != nil {
		http.Error(w, "failed to update cart", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cart)
}
