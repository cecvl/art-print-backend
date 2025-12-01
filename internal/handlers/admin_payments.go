package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/payment"
	"github.com/cecvl/art-print-backend/internal/services/payment/providers"
)

// GetAdminPaymentsHandler lists payments with optional filters
func GetAdminPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := r.URL.Query().Get("status")
	orderId := r.URL.Query().Get("orderId")
	buyerId := r.URL.Query().Get("buyerId")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	var docs []*firestore.DocumentSnapshot
	var err error

	if orderId != "" {
		docs, err = firebase.FirestoreClient.Collection("payments").Where("orderId", "==", orderId).OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
	} else if buyerId != "" {
		docs, err = firebase.FirestoreClient.Collection("payments").Where("buyerId", "==", buyerId).OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
	} else if status != "" {
		docs, err = firebase.FirestoreClient.Collection("payments").Where("status", "==", status).OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
	} else {
		docs, err = firebase.FirestoreClient.Collection("payments").OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx).GetAll()
	}

	if err != nil {
		log.Printf("❌ failed to query payments: %v", err)
		http.Error(w, "failed to query payments", http.StatusInternalServerError)
		return
	}

	out := make([]*models.Payment, 0, len(docs))
	for _, d := range docs {
		var p models.Payment
		if err := d.DataTo(&p); err != nil {
			continue
		}
		out = append(out, &p)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"payments": out})
}

// GetAdminPaymentHandler returns a single payment
func GetAdminPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	paymentId := r.URL.Query().Get("paymentId")
	if paymentId == "" {
		http.Error(w, "paymentId required", http.StatusBadRequest)
		return
	}

	repo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	p, err := repo.GetPaymentByID(ctx, paymentId)
	if err != nil {
		http.Error(w, "payment not found", http.StatusNotFound)
		return
	}

	// include linked order if exists
	var order *models.Order
	if p.OrderID != "" {
		if doc, err := firebase.FirestoreClient.Collection("orders").Doc(p.OrderID).Get(ctx); err == nil {
			var o models.Order
			if err := doc.DataTo(&o); err == nil {
				order = &o
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"payment": p, "order": order})
}

type verifyReq struct {
	PaymentID string `json:"paymentId"`
}

// VerifyPaymentAdminHandler verifies payment status with provider and updates records
func VerifyPaymentAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body verifyReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.PaymentID == "" {
		http.Error(w, "paymentId required", http.StatusBadRequest)
		return
	}

	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	provider := providers.NewSimulatedProvider()
	paymentService := payment.NewPaymentService(paymentRepo, provider)

	p, err := paymentService.VerifyPayment(ctx, body.PaymentID)
	if err != nil {
		log.Printf("❌ verify payment failed: %v", err)
		http.Error(w, "failed to verify payment", http.StatusInternalServerError)
		return
	}

	writeAdminAction(ctx, r, "verify_payment", "payment", body.PaymentID, nil)

	_ = json.NewEncoder(w).Encode(p)
}

type paymentRefundReq struct {
	PaymentID string  `json:"paymentId"`
	Amount    float64 `json:"amount,omitempty"`
	Reason    string  `json:"reason,omitempty"`
}

// RefundPaymentAdminHandler triggers provider refund and updates records
func RefundPaymentAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body paymentRefundReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.PaymentID == "" {
		http.Error(w, "paymentId required", http.StatusBadRequest)
		return
	}

	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	provider := providers.NewSimulatedProvider()
	paymentService := payment.NewPaymentService(paymentRepo, provider)

	if err := paymentService.RefundPayment(ctx, body.PaymentID, body.Amount); err != nil {
		log.Printf("❌ refund failed: %v", err)
		http.Error(w, "failed to process refund", http.StatusInternalServerError)
		return
	}

	writeAdminAction(ctx, r, "refund_payment", "payment", body.PaymentID, map[string]interface{}{"amount": body.Amount, "reason": body.Reason})

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "refunded", "paymentId": body.PaymentID})
}
