package providers

import (
	"context"
)

// PaymentProvider defines the interface for payment providers
type PaymentProvider interface {
	// CreatePayment initiates a payment and returns a transaction ID
	CreatePayment(ctx context.Context, amount float64, orderID string, metadata map[string]string) (string, error)

	// VerifyPayment verifies the status of a payment
	VerifyPayment(ctx context.Context, transactionID string) (bool, error)

	// RefundPayment processes a refund
	RefundPayment(ctx context.Context, transactionID string, amount float64) error

	// GetProviderName returns the name of the provider
	GetProviderName() string
}

