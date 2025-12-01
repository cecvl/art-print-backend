package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/payment"
	"github.com/cecvl/art-print-backend/internal/services/payment/providers"
)

// guardDev ensures dev-only endpoints are disabled in production
func guardDev() bool {
	env := os.Getenv("APP_ENV")
	if env == "dev" {
		return true
	}
	// Allow override for CI/dev convenience
	if os.Getenv("ADMIN_DEV_ALLOW") == "true" {
		return true
	}
	return false
}

// AddServicesHandler creates service docs under a shop
// Body: { "shopId": "...", "services": [ { name, description, technology, basePrice } ] }
func AddServicesHandler(w http.ResponseWriter, r *http.Request) {
	if !guardDev() {
		http.Error(w, "dev endpoints disabled", http.StatusForbidden)
		return
	}
	ctx := r.Context()
	var body struct {
		ShopID   string                `json:"shopId"`
		Services []models.PrintService `json:"services"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ShopID == "" || len(body.Services) == 0 {
		http.Error(w, "shopId and services required", http.StatusBadRequest)
		return
	}

	repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
	created := []string{}
	for _, s := range body.Services {
		s.ID = uuid.NewString()
		s.ShopID = body.ShopID
		s.CreatedAt = time.Now()
		s.UpdatedAt = time.Now()
		s.IsActive = true
		if err := repo.CreateService(ctx, &s); err != nil {
			log.Printf("⚠️ failed to create service: %v", err)
			continue
		}
		// write service change entry
		_, _, _ = firebase.FirestoreClient.Collection("services_changes").Add(ctx, map[string]interface{}{"serviceId": s.ID, "shopId": s.ShopID, "action": "created", "details": s, "createdAt": time.Now(), "createdBy": ctx.Value("userId")})
		created = append(created, s.ID)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"created": created})
}

// SimulateOrdersHandler creates synthetic orders over past N months
// Body: { "months": 3, "perMonth": 50, "shopId": "optional" }
func SimulateOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if !guardDev() {
		http.Error(w, "dev endpoints disabled", http.StatusForbidden)
		return
	}
	ctx := r.Context()
	var body struct {
		Months   int    `json:"months"`
		PerMonth int    `json:"perMonth"`
		ShopID   string `json:"shopId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Months <= 0 {
		body.Months = 3
	}
	if body.PerMonth <= 0 {
		body.PerMonth = 10
	}

	// gather sample artworks
	artsDocs, _ := firebase.FirestoreClient.Collection("artworks").Limit(200).Documents(ctx).GetAll()
	artworkIDs := []string{}
	for _, d := range artsDocs {
		var a models.Artwork
		if d.DataTo(&a) == nil {
			artworkIDs = append(artworkIDs, d.Ref.ID)
		}
	}
	if len(artworkIDs) == 0 {
		http.Error(w, "no artworks found to simulate orders", http.StatusBadRequest)
		return
	}

	// pick a printshop if not provided
	shopId := body.ShopID
	if shopId == "" {
		shopsDocs, _ := firebase.FirestoreClient.Collection("printshops").Where("isActive", "==", true).Limit(10).Documents(ctx).GetAll()
		if len(shopsDocs) > 0 {
			shopId = shopsDocs[rand.Intn(len(shopsDocs))].Ref.ID
		}
	}

	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	provider := providers.NewSimulatedProvider()
	paymentService := payment.NewPaymentService(paymentRepo, provider)

	createdOrders := []string{}
	rand.Seed(time.Now().UnixNano())
	for m := 0; m < body.Months; m++ {
		for i := 0; i < body.PerMonth; i++ {
			// choose artwork
			aid := artworkIDs[rand.Intn(len(artworkIDs))]
			qty := 1 + rand.Intn(3)
			price := 20 + rand.Float64()*180

			createdAt := time.Now().AddDate(0, -m, -rand.Intn(27))

			order := models.Order{
				OrderID:     uuid.NewString(),
				BuyerID:     "sim_buyer_" + uuid.NewString(),
				PrintShopID: shopId,
				Items: []models.CartItem{
					{ArtworkID: aid, Quantity: qty, Price: price},
				},
				TotalAmount:   float64(qty) * price,
				PaymentMethod: "simulated",
				PaymentStatus: string(models.PaymentStatusCompleted),
				Status:        "completed",
				CreatedAt:     createdAt,
				UpdatedAt:     createdAt,
			}

			// persist order
			_, err := firebase.FirestoreClient.Collection("orders").Doc(order.OrderID).Set(ctx, order)
			if err != nil {
				log.Printf("⚠️ failed to write simulated order: %v", err)
				continue
			}

			// create payment record
			payReq := models.PaymentRequest{OrderID: order.OrderID, Amount: order.TotalAmount, PaymentMethod: "simulated", PaymentType: "full", Metadata: map[string]string{"simulated": "true"}}
			pmt, err := paymentService.CreatePayment(ctx, payReq, order.TotalAmount)
			if err == nil && pmt != nil {
				// link payment id to order
				_, _ = firebase.FirestoreClient.Collection("orders").Doc(order.OrderID).Update(ctx, []firestore.Update{{Path: "paymentId", Value: pmt.ID}, {Path: "paymentStatus", Value: string(models.PaymentStatusCompleted)}})
			}

			createdOrders = append(createdOrders, order.OrderID)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"created": len(createdOrders), "orderIds": createdOrders})
}
