package providers

import (
	"context"
	"fmt"
	"time"
)

// SimulatedProvider simulates payment processing for testing
type SimulatedProvider struct {
	// In-memory store for simulated transactions
	transactions map[string]*SimulatedTransaction
}

// SimulatedTransaction represents a simulated payment transaction
type SimulatedTransaction struct {
	TransactionID string
	Amount        float64
	OrderID       string
	Status        string // "pending", "completed", "failed"
	CreatedAt     time.Time
}

// NewSimulatedProvider creates a new simulated payment provider
func NewSimulatedProvider() *SimulatedProvider {
	return &SimulatedProvider{
		transactions: make(map[string]*SimulatedTransaction),
	}
}

// GetProviderName returns the provider name
func (p *SimulatedProvider) GetProviderName() string {
	return "simulated"
}

// CreatePayment simulates creating a payment
// In simulation, payments are automatically completed after a short delay
func (p *SimulatedProvider) CreatePayment(ctx context.Context, amount float64, orderID string, metadata map[string]string) (string, error) {
	transactionID := fmt.Sprintf("sim_%d_%s", time.Now().Unix(), orderID)

	transaction := &SimulatedTransaction{
		TransactionID: transactionID,
		Amount:        amount,
		OrderID:       orderID,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	p.transactions[transactionID] = transaction

	// In simulation, automatically complete payment after 1 second
	// In real implementation, this would be handled by webhook
	go func() {
		time.Sleep(1 * time.Second)
		if txn, exists := p.transactions[transactionID]; exists {
			txn.Status = "completed"
		}
	}()

	return transactionID, nil
}

// VerifyPayment verifies the status of a simulated payment
func (p *SimulatedProvider) VerifyPayment(ctx context.Context, transactionID string) (bool, error) {
	transaction, exists := p.transactions[transactionID]
	if !exists {
		return false, fmt.Errorf("transaction not found: %s", transactionID)
	}

	return transaction.Status == "completed", nil
}

// RefundPayment simulates a refund
func (p *SimulatedProvider) RefundPayment(ctx context.Context, transactionID string, amount float64) error {
	transaction, exists := p.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction not found: %s", transactionID)
	}

	if transaction.Status != "completed" {
		return fmt.Errorf("cannot refund transaction with status: %s", transaction.Status)
	}

	// In simulation, mark as refunded
	transaction.Status = "refunded"
	return nil
}

// GetTransaction retrieves a simulated transaction (for testing)
func (p *SimulatedProvider) GetTransaction(transactionID string) (*SimulatedTransaction, bool) {
	txn, exists := p.transactions[transactionID]
	return txn, exists
}
