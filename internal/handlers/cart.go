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

// ===== Add item to cart =====
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	//Calculate total cost //From []CartItem in Order Struct
	var req struct {
		ArtworkID string `json:"artworkId"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate artwork & get trusted price
	artDoc, err := firebase.FirestoreClient.Collection("artworks").Doc(req.ArtworkID).Get(ctx)
	if err != nil || !artDoc.Exists() {
		http.Error(w, "Artwork not found", http.StatusNotFound)
		return
	}
	var art models.Artwork
	if err := artDoc.DataTo(&art); err != nil {
		http.Error(w, "Error parsing artwork", http.StatusInternalServerError)
		return
	}
	if !art.IsAvailable {
		http.Error(w, "Artwork not available", http.StatusConflict)
		return
	}
	price, ok := art.PrintOptions["basePrice"].(float64)
	if !ok {
		http.Error(w, "Invalid price data", http.StatusInternalServerError)
		return
	}

	// Find or create cart order
	var cartOrder *firestore.DocumentSnapshot
	iter := firebase.FirestoreClient.Collection("orders").
		Where("buyerId", "==", userID).
		Where("status", "==", "cart").
		Limit(1).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Error checking cart", http.StatusInternalServerError)
			return
		}
		cartOrder = doc
	}
	iter.Stop()

	var order models.Order
	if cartOrder != nil {
		if err := cartOrder.DataTo(&order); err != nil {
			http.Error(w, "Error reading cart order", http.StatusInternalServerError)
			return
		}
	} else {
		order = models.Order{
			BuyerID:   userID,
			Items:     []models.CartItem{},
			Status:    "cart",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Add or update cart item
	itemExists := false
	for i, item := range order.Items {
		if item.ArtworkID == req.ArtworkID {
			order.Items[i].Quantity += req.Quantity
			itemExists = true
			break
		}
	}
	if !itemExists {
		order.Items = append(order.Items, models.CartItem{
			ArtworkID: req.ArtworkID,
			Quantity:  req.Quantity,
			Price:     price,
		})
	}

	// Recalculate total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total
	order.UpdatedAt = time.Now()

	// Save cart
	if cartOrder != nil {
		_, err = cartOrder.Ref.Set(ctx, order)
	} else {
		_, _, err = firebase.FirestoreClient.Collection("orders").Add(ctx, order)
	}
	if err != nil {
		http.Error(w, "Failed to save cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// ===== View cart =====
func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	iter := firebase.FirestoreClient.Collection("orders").
		Where("buyerId", "==", userID).
		Where("status", "==", "cart").
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "Cart is empty", http.StatusNotFound)
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

	json.NewEncoder(w).Encode(order)
}

// ===== Remove item from cart =====
func RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)
	artworkID := r.URL.Query().Get("artworkId")
	if artworkID == "" {
		http.Error(w, "Missing artworkId", http.StatusBadRequest)
		return
	}

	iter := firebase.FirestoreClient.Collection("orders").
		Where("buyerId", "==", userID).
		Where("status", "==", "cart").
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "Cart is empty", http.StatusNotFound)
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

	// Filter out removed item
	newItems := []models.CartItem{}
	for _, item := range order.Items {
		if item.ArtworkID != artworkID {
			newItems = append(newItems, item)
		}
	}
	order.Items = newItems

	// Recalculate total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total
	order.UpdatedAt = time.Now()

	_, err = doc.Ref.Set(ctx, order)
	if err != nil {
		http.Error(w, "Failed to update cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}
