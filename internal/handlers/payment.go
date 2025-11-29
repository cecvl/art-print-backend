package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/payment"
	"github.com/cecvl/art-print-backend/internal/services/payment/providers"
)

// PaymentHandler handles payment-related requests
type PaymentHandler struct {
	paymentService *payment.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler() *PaymentHandler {
	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	simulatedProvider := providers.NewSimulatedProvider()
	paymentService := payment.NewPaymentService(paymentRepo, simulatedProvider)

	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentHandler creates a payment for an order
func (h *PaymentHandler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buyerID := ctx.Value("userID").(string)

	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get order to verify ownership and get amount
	fsClient := firebase.FirestoreClient
	orderDoc, err := fsClient.Collection("orders").Doc(req.OrderID).Get(ctx)
	if err != nil {
		log.Printf("❌ Order not found: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	var order models.Order
	if err := orderDoc.DataTo(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusInternalServerError)
		return
	}

	// Verify order ownership
	if order.BuyerID != buyerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Create payment
	payment, err := h.paymentService.CreatePayment(ctx, req, order.TotalAmount)
	if err != nil {
		log.Printf("❌ Failed to create payment: %v", err)
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}

	// Set buyer ID
	payment.BuyerID = buyerID
	// Update payment with buyer ID
	paymentRepo := repositories.NewPaymentRepository(firebase.FirestoreClient)
	if err := paymentRepo.UpdatePayment(ctx, payment.ID, map[string]interface{}{
		"buyerId":  buyerID,
		"updatedAt": time.Now(),
	}); err != nil {
		log.Printf("⚠️ Failed to update buyer ID: %v", err)
	}

	// Update order with payment ID
	order.PaymentID = payment.ID
	order.UpdatedAt = time.Now()
	if _, err := fsClient.Collection("orders").Doc(req.OrderID).Set(ctx, order); err != nil {
		log.Printf("⚠️ Failed to update order with payment ID: %v", err)
	}

	// For simulated provider, automatically verify after creation
	if req.PaymentMethod == "simulated" {
		go func() {
			time.Sleep(2 * time.Second) // Wait for simulated payment to complete
			verifiedPayment, err := h.paymentService.VerifyPayment(ctx, payment.ID)
			if err == nil && verifiedPayment.Status == models.PaymentStatusCompleted {
				// Update order status
				order.Status = "confirmed"
				paymentStatus, _, _ := h.paymentService.CalculatePaymentStatus(ctx, req.OrderID, order.TotalAmount)
				order.PaymentStatus = paymentStatus
				order.UpdatedAt = time.Now()
				fsClient.Collection("orders").Doc(req.OrderID).Set(ctx, order)
				log.Printf("✅ Order %s confirmed after payment", req.OrderID)
			}
		}()
	}

	response := models.PaymentResponse{
		PaymentID:     payment.ID,
		OrderID:       payment.OrderID,
		Amount:         payment.Amount,
		Status:         payment.Status,
		TransactionID: payment.TransactionID,
		ProviderData:   payment.ProviderData,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// VerifyPaymentHandler verifies a payment status
func (h *PaymentHandler) VerifyPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buyerID := ctx.Value("userID").(string)

	// Extract payment ID from URL
	paymentID := r.URL.Query().Get("id")
	if paymentID == "" {
		http.Error(w, "Payment ID required", http.StatusBadRequest)
		return
	}

	// Get payment
	payment, err := h.paymentService.VerifyPayment(ctx, paymentID)
	if err != nil {
		log.Printf("❌ Failed to verify payment: %v", err)
		http.Error(w, "Payment not found or verification failed", http.StatusNotFound)
		return
	}

	// Verify ownership
	if payment.BuyerID != buyerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Update order status if payment completed
	if payment.Status == models.PaymentStatusCompleted {
		fsClient := firebase.FirestoreClient
		orderDoc, err := fsClient.Collection("orders").Doc(payment.OrderID).Get(ctx)
		if err == nil {
			var order models.Order
			if err := orderDoc.DataTo(&order); err == nil {
				order.Status = "confirmed"
				paymentStatus, _, _ := h.paymentService.CalculatePaymentStatus(ctx, payment.OrderID, order.TotalAmount)
				order.PaymentStatus = paymentStatus
				order.UpdatedAt = time.Now()
				fsClient.Collection("orders").Doc(payment.OrderID).Set(ctx, order)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// GetPaymentsHandler retrieves payments for an order
func (h *PaymentHandler) GetPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buyerID := ctx.Value("userID").(string)

	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	// Verify order ownership
	fsClient := firebase.FirestoreClient
	orderDoc, err := fsClient.Collection("orders").Doc(orderID).Get(ctx)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	var order models.Order
	if err := orderDoc.DataTo(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusInternalServerError)
		return
	}

	if order.BuyerID != buyerID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get payments
	payments, err := h.paymentService.GetPaymentHistory(ctx, orderID)
	if err != nil {
		log.Printf("❌ Failed to get payments: %v", err)
		http.Error(w, "Failed to get payments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// PaymentWebhookHandler handles webhooks from payment providers
func (h *PaymentHandler) PaymentWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var webhook models.PaymentWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	// Process webhook
	if err := h.paymentService.ProcessPaymentWebhook(ctx, webhook); err != nil {
		log.Printf("❌ Failed to process webhook: %v", err)
		http.Error(w, "Failed to process webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// RefundPaymentHandler processes a refund (admin only)
func (h *PaymentHandler) RefundPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		PaymentID string  `json:"paymentId"`
		Amount    float64 `json:"amount,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process refund
	if err := h.paymentService.RefundPayment(ctx, req.PaymentID, req.Amount); err != nil {
		log.Printf("❌ Failed to process refund: %v", err)
		http.Error(w, "Failed to process refund", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Refund processed successfully"})
}

