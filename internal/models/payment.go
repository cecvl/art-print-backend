package models

import "time"

// Payment represents a payment transaction
type Payment struct {
	ID            string                 `firestore:"id" json:"id"`
	OrderID       string                 `firestore:"orderId" json:"orderId"`
	BuyerID       string                 `firestore:"buyerId" json:"buyerId"`
	Amount        float64                `firestore:"amount" json:"amount"`
	PaymentMethod string                 `firestore:"paymentMethod" json:"paymentMethod"` // "stripe", "mpesa", "simulated"
	Status        PaymentStatus          `firestore:"status" json:"status"`               // "pending", "processing", "completed", "failed", "refunded"
	TransactionID string                 `firestore:"transactionId" json:"transactionId"` // External provider transaction ID
	PaymentType   PaymentType            `firestore:"paymentType" json:"paymentType"`     // "deposit" (50%), "full" (100%), "remaining" (remaining 50%)
	ProviderData  map[string]interface{} `firestore:"providerData" json:"providerData"`   // Provider-specific data
	CreatedAt     time.Time              `firestore:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time              `firestore:"updatedAt" json:"updatedAt"`
	CompletedAt   *time.Time             `firestore:"completedAt,omitempty" json:"completedAt,omitempty"`
	FailedAt      *time.Time             `firestore:"failedAt,omitempty" json:"failedAt,omitempty"`
	FailureReason string                 `firestore:"failureReason,omitempty" json:"failureReason,omitempty"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypeDeposit   PaymentType = "deposit"   // 50% initial payment
	PaymentTypeFull      PaymentType = "full"      // 100% full payment
	PaymentTypeRemaining PaymentType = "remaining" // Remaining 50% after deposit
)

// PaymentRequest represents a request to create a payment
type PaymentRequest struct {
	OrderID       string            `json:"orderId"`
	Amount        float64           `json:"amount,omitempty"`   // Optional: if not provided, calculated from order
	PaymentMethod string            `json:"paymentMethod"`      // "stripe", "mpesa", "simulated"
	PaymentType   string            `json:"paymentType"`        // "deposit", "full", "remaining"
	Metadata      map[string]string `json:"metadata,omitempty"` // Additional metadata for provider
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	PaymentID     string                 `json:"paymentId"`
	OrderID       string                 `json:"orderId"`
	Amount        float64                `json:"amount"`
	Status        PaymentStatus          `json:"status"`
	TransactionID string                 `json:"transactionId"`
	PaymentURL    string                 `json:"paymentUrl,omitempty"` // For providers that need redirect
	ProviderData  map[string]interface{} `json:"providerData,omitempty"`
}

// PaymentWebhook represents a webhook payload from payment provider
type PaymentWebhook struct {
	TransactionID string                 `json:"transactionId"`
	Status        string                 `json:"status"`
	Amount        float64                `json:"amount"`
	Metadata      map[string]interface{} `json:"metadata"`
	Signature     string                 `json:"signature,omitempty"` // For webhook verification
}
