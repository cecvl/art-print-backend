package repositories

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/models"
	"github.com/google/uuid"
)

// PaymentRepository handles payment data operations
type PaymentRepository struct {
	client *firestore.Client
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(client *firestore.Client) *PaymentRepository {
	return &PaymentRepository{client: client}
}

// CreatePayment creates a new payment record
func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.ID == "" {
		payment.ID = uuid.NewString()
	}
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
	}
	payment.UpdatedAt = time.Now()

	_, err := r.client.Collection("payments").Doc(payment.ID).Set(ctx, payment)
	return err
}

// GetPaymentByID retrieves a payment by its ID
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	doc, err := r.client.Collection("payments").Doc(paymentID).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, errors.New("payment not found")
	}

	var payment models.Payment
	if err := doc.DataTo(&payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentsByOrderID retrieves all payments for an order
func (r *PaymentRepository) GetPaymentsByOrderID(ctx context.Context, orderID string) ([]*models.Payment, error) {
	docs, err := r.client.Collection("payments").
		Where("orderId", "==", orderID).
		OrderBy("createdAt", firestore.Asc).
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	payments := make([]*models.Payment, 0, len(docs))
	for _, doc := range docs {
		var payment models.Payment
		if err := doc.DataTo(&payment); err != nil {
			continue
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// GetPaymentsByBuyerID retrieves all payments for a buyer
func (r *PaymentRepository) GetPaymentsByBuyerID(ctx context.Context, buyerID string) ([]*models.Payment, error) {
	docs, err := r.client.Collection("payments").
		Where("buyerId", "==", buyerID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	payments := make([]*models.Payment, 0, len(docs))
	for _, doc := range docs {
		var payment models.Payment
		if err := doc.DataTo(&payment); err != nil {
			continue
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// UpdatePayment updates a payment record
func (r *PaymentRepository) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	updatesList := make([]firestore.Update, 0, len(updates))
	for field, value := range updates {
		updatesList = append(updatesList, firestore.Update{
			Path:  field,
			Value: value,
		})
	}

	_, err := r.client.Collection("payments").Doc(paymentID).Update(ctx, updatesList)
	return err
}

// UpdatePaymentStatus updates the payment status
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error {
	updates := map[string]interface{}{
		"status":    status,
		"updatedAt": time.Now(),
	}

	if status == models.PaymentStatusCompleted {
		now := time.Now()
		updates["completedAt"] = now
	} else if status == models.PaymentStatusFailed {
		now := time.Now()
		updates["failedAt"] = now
	}

	return r.UpdatePayment(ctx, paymentID, updates)
}

// GetPaymentsByStatus retrieves payments by status
func (r *PaymentRepository) GetPaymentsByStatus(ctx context.Context, status models.PaymentStatus) ([]*models.Payment, error) {
	docs, err := r.client.Collection("payments").
		Where("status", "==", status).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	payments := make([]*models.Payment, 0, len(docs))
	for _, doc := range docs {
		var payment models.Payment
		if err := doc.DataTo(&payment); err != nil {
			continue
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}
