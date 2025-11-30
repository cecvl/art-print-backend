package payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/cecvl/art-print-backend/internal/repositories"
	"github.com/cecvl/art-print-backend/internal/services/payment/providers"
)

// PaymentService handles payment operations
type PaymentService struct {
	repo     *repositories.PaymentRepository
	provider providers.PaymentProvider
}

// NewPaymentService creates a new payment service
func NewPaymentService(repo *repositories.PaymentRepository, provider providers.PaymentProvider) *PaymentService {
	return &PaymentService{
		repo:     repo,
		provider: provider,
	}
}

// CreatePayment creates a new payment for an order
func (s *PaymentService) CreatePayment(ctx context.Context, req models.PaymentRequest, orderAmount float64) (*models.Payment, error) {
	// Calculate payment amount based on type
	var amount float64
	var paymentType models.PaymentType

	switch req.PaymentType {
	case "deposit":
		amount = orderAmount * 0.5 // 50% deposit
		paymentType = models.PaymentTypeDeposit
	case "full":
		amount = orderAmount // 100% full payment
		paymentType = models.PaymentTypeFull
	case "remaining":
		amount = orderAmount * 0.5 // Remaining 50%
		paymentType = models.PaymentTypeRemaining
	default:
		// Default to deposit if not specified
		amount = orderAmount * 0.5
		paymentType = models.PaymentTypeDeposit
	}

	// Override with explicit amount if provided
	if req.Amount > 0 {
		amount = req.Amount
	}

	// Create payment with provider
	transactionID, err := s.provider.CreatePayment(ctx, amount, req.OrderID, req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment with provider: %w", err)
	}

	// Create payment record
	payment := &models.Payment{
		ID:            "",
		OrderID:       req.OrderID,
		BuyerID:       "", // Will be set from order
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		Status:        models.PaymentStatusPending,
		TransactionID: transactionID,
		PaymentType:   paymentType,
		ProviderData:  make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save payment
	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	log.Printf("✅ Created payment %s for order %s: %.2f", payment.ID, req.OrderID, amount)
	return payment, nil
}

// ProcessPaymentWebhook processes a webhook from payment provider
func (s *PaymentService) ProcessPaymentWebhook(ctx context.Context, webhook models.PaymentWebhook) error {
	// Find payment by transaction ID
	// Note: In real implementation, we'd need to store transactionID -> paymentID mapping
	// For now, we'll search by transactionID in provider data or use a separate index

	// Verify payment with provider
	verified, err := s.provider.VerifyPayment(ctx, webhook.TransactionID)
	if err != nil {
		return fmt.Errorf("failed to verify payment: %w", err)
	}

	if !verified {
		return fmt.Errorf("payment verification failed for transaction: %s", webhook.TransactionID)
	}

	// Update payment status
	// In real implementation, we'd find the payment by transactionID
	// For now, this is a placeholder - we'll need to add a method to find by transactionID
	log.Printf("✅ Payment verified for transaction: %s", webhook.TransactionID)

	return nil
}

// VerifyPayment verifies a payment status
func (s *PaymentService) VerifyPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Verify with provider
	verified, err := s.provider.VerifyPayment(ctx, payment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify payment: %w", err)
	}

	if verified && payment.Status != models.PaymentStatusCompleted {
		// Update payment status to completed
		now := time.Now()
		updates := map[string]interface{}{
			"status":      models.PaymentStatusCompleted,
			"completedAt": now,
			"updatedAt":   now,
		}

		if err := s.repo.UpdatePayment(ctx, paymentID, updates); err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		payment.Status = models.PaymentStatusCompleted
		completedAt := now
		payment.CompletedAt = &completedAt
	}

	return payment, nil
}

// RefundPayment processes a refund
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount float64) error {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.PaymentStatusCompleted {
		return fmt.Errorf("cannot refund payment with status: %s", payment.Status)
	}

	// Process refund with provider
	refundAmount := amount
	if refundAmount == 0 {
		refundAmount = payment.Amount // Full refund if amount not specified
	}

	if err := s.provider.RefundPayment(ctx, payment.TransactionID, refundAmount); err != nil {
		return fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment status
	updates := map[string]interface{}{
		"status":    models.PaymentStatusRefunded,
		"updatedAt": time.Now(),
	}

	if err := s.repo.UpdatePayment(ctx, paymentID, updates); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	log.Printf("✅ Refunded payment %s: %.2f", paymentID, refundAmount)
	return nil
}

// GetPaymentHistory retrieves payment history for an order
func (s *PaymentService) GetPaymentHistory(ctx context.Context, orderID string) ([]*models.Payment, error) {
	return s.repo.GetPaymentsByOrderID(ctx, orderID)
}

// CalculatePaymentStatus calculates the overall payment status for an order
func (s *PaymentService) CalculatePaymentStatus(ctx context.Context, orderID string, orderAmount float64) (string, float64, error) {
	payments, err := s.repo.GetPaymentsByOrderID(ctx, orderID)
	if err != nil {
		return "unpaid", 0, err
	}

	var totalPaid float64
	for _, payment := range payments {
		if payment.Status == models.PaymentStatusCompleted {
			totalPaid += payment.Amount
		}
	}

	if totalPaid == 0 {
		return "unpaid", 0, nil
	} else if totalPaid >= orderAmount {
		return "paid", totalPaid, nil
	} else {
		return "partial", totalPaid, nil
	}
}
