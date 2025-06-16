package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/models"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	userDoc, err := firebase.FirestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil || userDoc.Data()["userType"] != models.Buyer {
		http.Error(w, "Only buyers can create orders", http.StatusForbidden)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return
	}

	printShops, err := firebase.FirestoreClient.Collection("users").
		Where("userType", "==", models.PrintShop).
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

