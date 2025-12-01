package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
)

// GetAdminPrintShopsHandler lists print shops (filter: isActive)
func GetAdminPrintShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	active := r.URL.Query().Get("isActive")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)

	var shops []*models.PrintShopProfile
	var err error
	if active == "true" {
		shops, err = repo.GetActiveShops(ctx)
	} else {
		// fetch all shops limited
		docs, dErr := firebase.FirestoreClient.Collection("printshops").OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
		if dErr != nil {
			log.Printf("❌ failed to query printshops: %v", dErr)
			http.Error(w, "failed to query printshops", http.StatusInternalServerError)
			return
		}
		for _, d := range docs {
			var s models.PrintShopProfile
			if err := d.DataTo(&s); err == nil {
				s.ID = d.Ref.ID
				shops = append(shops, &s)
			}
		}
		err = nil
	}

	if err != nil {
		log.Printf("❌ failed to fetch shops: %v", err)
		http.Error(w, "failed to fetch shops", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"printshops": shops})
}

// GetAdminPrintShopHandler returns a shop profile and services
func GetAdminPrintShopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopId := r.URL.Query().Get("shopId")
	if shopId == "" {
		http.Error(w, "shopId required", http.StatusBadRequest)
		return
	}

	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	shop, err := repo.GetShopByID(ctx, shopId)
	if err != nil {
		http.Error(w, "shop not found", http.StatusNotFound)
		return
	}

	services, _ := repo.GetServicesByShopID(ctx, shopId)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"shop": shop, "services": services})
}

type updateServicePriceReq struct {
	ServiceID string  `json:"serviceId"`
	Price     float64 `json:"price"`
}

// UpdateServicePriceHandler updates a service's base price (and records admin action)
func UpdateServicePriceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body updateServicePriceReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ServiceID == "" {
		http.Error(w, "serviceId required", http.StatusBadRequest)
		return
	}

	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	updates := map[string]interface{}{"price": body.Price, "updatedAt": time.Now()}
	if err := repo.UpdateService(ctx, body.ServiceID, updates); err != nil {
		log.Printf("❌ failed to update service price: %v", err)
		http.Error(w, "failed to update service", http.StatusInternalServerError)
		return
	}

	writeAdminAction(ctx, r, "update_service_price", "service", body.ServiceID, map[string]interface{}{"price": body.Price})

	w.WriteHeader(http.StatusNoContent)
}
