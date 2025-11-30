package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/services/catalog"
	"github.com/cecvl/art-print-backend/internal/services/pricing"
)

// printOptionsEqual compares two PrintOrderOptions for equality
func printOptionsEqual(a, b models.PrintOrderOptions) bool {
	return reflect.DeepEqual(a, b)
}

// AddToCartHandler adds or updates an item in the user's cart
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fsClient := firebase.GetFirestoreClient()
	uid := ctx.Value("userId")
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	buyerID := uid.(string)

	var newItem models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate price using legacy catalog pricing if price is not provided
	if newItem.Price == 0 {
		catSvc := catalog.NewCatalogService()
		priceSvc := pricing.NewPricingService()
		req := pricing.PriceRequest{
			Size:      newItem.PrintOptions.Size,
			Frame:     newItem.PrintOptions.Frame,
			Material:  newItem.PrintOptions.Material,
			Medium:    newItem.PrintOptions.Medium,
			Quantity:  newItem.Quantity,
			RushOrder: newItem.PrintOptions.RushOrder,
		}
		opts := catSvc.GetPrintOptions()
		resp := priceSvc.Calculate(req, opts)
		newItem.Price = float64(resp.Total)
	}

	cartRef := fsClient.Collection("carts").Doc(buyerID)
	doc, err := cartRef.Get(ctx)

	var cart models.Cart
	if err == nil && doc.Exists() {
		_ = doc.DataTo(&cart)
		updated := false
		for i, item := range cart.Items {
			// Consider items identical only if artwork AND print options match
			if item.ArtworkID == newItem.ArtworkID && printOptionsEqual(item.PrintOptions, newItem.PrintOptions) {
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
	ctx := r.Context()
	fsClient := firebase.GetFirestoreClient()
	uid := ctx.Value("userId")
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	buyerID := uid.(string)

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
	ctx := r.Context()
	fsClient := firebase.GetFirestoreClient()
	uid := ctx.Value("userId")
	if uid == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	buyerID := uid.(string)

	var payload struct {
		ArtworkID    string                   `json:"artworkId"`
		PrintOptions models.PrintOrderOptions `json:"printOptions,omitempty"`
		UseOptions   bool                     `json:"useOptions"`
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
		// If UseOptions is true, only remove items that match both artwork and print options
		if payload.UseOptions {
			if item.ArtworkID == payload.ArtworkID && printOptionsEqual(item.PrintOptions, payload.PrintOptions) {
				continue // skip (remove) this specific variant
			}
			newItems = append(newItems, item)
		} else {
			// Remove all items with the artwork id
			if item.ArtworkID == payload.ArtworkID {
				continue
			}
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
