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
	"github.com/cecvl/art-print-backend/internal/services/payment"
	"github.com/cecvl/art-print-backend/internal/services/payment/providers"
)

// GetAdminOrdersHandler lists orders with optional filters
func GetAdminOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := firebase.FirestoreClient.Collection("orders").OrderBy("createdAt", firestore.Desc)

	// filters
	status := r.URL.Query().Get("status")
	buyerId := r.URL.Query().Get("buyerId")
	shopId := r.URL.Query().Get("printShopId")
	createdAfter := r.URL.Query().Get("createdAfter")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	if status != "" {
		q = q.Where("status", "==", status)
	}
	if buyerId != "" {
		q = q.Where("buyerId", "==", buyerId)
	}
	if shopId != "" {
		q = q.Where("printShopId", "==", shopId)
	}
	if createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			q = q.Where("createdAt", ">=", t)
		}
	}

	q = q.OrderBy("createdAt", firestore.Desc).Limit(limit)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, "failed to query orders", http.StatusInternalServerError)
		return
	}

	out := make([]*models.Order, 0, len(docs))
	for _, d := range docs {
		var o models.Order
		if err := d.DataTo(&o); err != nil {
			log.Printf("⚠️ failed to parse order: %v", err)
			continue
		}
		out = append(out, &o)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"orders": out})
}

// GetAdminOrderHandler returns a single order with payments
func GetAdminOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderId := r.URL.Query().Get("orderId")
	if orderId == "" {
		http.Error(w, "missing orderId", http.StatusBadRequest)
		return
	}

	doc, err := firebase.FirestoreClient.Collection("orders").Doc(orderId).Get(ctx)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		http.Error(w, "invalid order data", http.StatusInternalServerError)
		return
	}

	// get payments
	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	payments, _ := paymentRepo.GetPaymentsByOrderID(ctx, orderId)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"order": order, "payments": payments})
}

type updateStatusReq struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
	Note    string `json:"note,omitempty"`
}

// UpdateOrderStatusHandler updates order status and writes an audit note
func UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body updateStatusReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.OrderID == "" || body.Status == "" {
		http.Error(w, "orderId and status required", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"status": body.Status, "updatedAt": time.Now()}
	if _, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to update order", http.StatusInternalServerError)
		return
	}

	// append admin note
	if body.Note != "" {
		note := map[string]interface{}{"note": body.Note, "createdAt": time.Now(), "createdBy": ctx.Value("userId")}
		_, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Update(ctx, []firestore.Update{{Path: "adminNotes", Value: firestore.ArrayUnion(note)}})
		if err != nil {
			log.Printf("⚠️ failed to append admin note: %v", err)
		}
	}

	writeAdminAction(ctx, r, "update_order_status", "order", body.OrderID, map[string]interface{}{"status": body.Status, "note": body.Note})

	w.WriteHeader(http.StatusNoContent)
}

type reassignReq struct {
	OrderID     string `json:"orderId"`
	PrintShopID string `json:"printShopId"`
}

// ReassignOrderHandler forces order assignment to a print shop
func ReassignOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body reassignReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.OrderID == "" || body.PrintShopID == "" {
		http.Error(w, "orderId and printShopId required", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"printShopId": body.PrintShopID, "updatedAt": time.Now()}
	if _, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to reassign order", http.StatusInternalServerError)
		return
	}

	// record assignment
	_, _, _ = firebase.FirestoreClient.Collection("assignments").Add(ctx, map[string]interface{}{"orderId": body.OrderID, "printShopId": body.PrintShopID, "status": "assigned", "createdAt": time.Now(), "createdBy": ctx.Value("userId")})

	writeAdminAction(ctx, r, "reassign_printshop", "order", body.OrderID, map[string]interface{}{"printShopId": body.PrintShopID})

	w.WriteHeader(http.StatusNoContent)
}

type cancelReq struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason,omitempty"`
}

// CancelOrderHandler marks an order cancelled
func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body cancelReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.OrderID == "" {
		http.Error(w, "orderId required", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"status": "cancelled", "updatedAt": time.Now()}
	if _, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to cancel order", http.StatusInternalServerError)
		return
	}

	// write admin note
	note := map[string]interface{}{"note": "cancelled: " + body.Reason, "createdAt": time.Now(), "createdBy": ctx.Value("userId")}
	_, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Update(ctx, []firestore.Update{{Path: "adminNotes", Value: firestore.ArrayUnion(note)}})
	if err != nil {
		log.Printf("⚠️ failed to append admin note: %v", err)
	}

	writeAdminAction(ctx, r, "cancel_order", "order", body.OrderID, map[string]interface{}{"reason": body.Reason})

	w.WriteHeader(http.StatusNoContent)
}

type refundReq struct {
	OrderID   string  `json:"orderId,omitempty"`
	PaymentID string  `json:"paymentId,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Reason    string  `json:"reason,omitempty"`
}

// RefundOrderHandler processes refunds either by paymentId or for all completed payments on an order
func RefundOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body refundReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	provider := providers.NewSimulatedProvider()
	paymentService := payment.NewPaymentService(paymentRepo, provider)

	var targets []string
	if body.PaymentID != "" {
		targets = append(targets, body.PaymentID)
	} else if body.OrderID != "" {
		payments, err := paymentRepo.GetPaymentsByOrderID(ctx, body.OrderID)
		if err != nil {
			http.Error(w, "failed to find payments", http.StatusInternalServerError)
			return
		}
		for _, p := range payments {
			if p.Status == models.PaymentStatusCompleted {
				targets = append(targets, p.ID)
			}
		}
	} else {
		http.Error(w, "paymentId or orderId required", http.StatusBadRequest)
		return
	}

	if len(targets) == 0 {
		http.Error(w, "no refundable payments found", http.StatusBadRequest)
		return
	}

	for _, pid := range targets {
		if err := paymentService.RefundPayment(ctx, pid, body.Amount); err != nil {
			log.Printf("❌ refund failed for %s: %v", pid, err)
			http.Error(w, "failed to process refund", http.StatusInternalServerError)
			return
		}
	}

	writeAdminAction(ctx, r, "refund_payment", "order", body.OrderID, map[string]interface{}{"paymentIds": targets, "amount": body.Amount, "reason": body.Reason})

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"refunded": targets})
}

// admin audit helper is in admin_audit.go
