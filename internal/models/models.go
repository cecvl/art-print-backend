package models

import "time"

const (
	Buyer     = "buyer"
	Artist    = "artist"
	PrintShop = "printShop"
)

type User struct {
	UID           string    `firestore:"uid"`           // Firebase UID
	Email         string    `firestore:"email"`         // Email address
	Roles         []string  `firestore:"roles"`         // ["buyer", "artist", "printShop"]
	Name          string    `firestore:"name"`          // Display name
	DateOfBirth   string    `firestore:"dateOfBirth"`   // Format: YYYY-MM-DD
	Description   string    `firestore:"description"`   // Profile bio
	AvatarURL     string    `firestore:"avatarUrl"`     // Cloudinary avatar image
	BackgroundURL string    `firestore:"backgroundUrl"` // Cloudinary cover image REMOVE FILE
	CreatedAt     time.Time `firestore:"createdAt"`     // Account creation time
}

type Artwork struct {
	ID           string                 `json:"id,omitempty"`
	ArtistID     string                 `firestore:"artistId"`
	Title        string                 `firestore:"title"`
	Description  string                 `firestore:"description"`
	ImageURL     string                 `firestore:"imageUrl" json:"imageUrl"`
	PrintOptions map[string]interface{} `firestore:"printOptions"`
	IsAvailable  bool                   `firestore:"isAvailable"`
	CreatedAt    time.Time              `firestore:"createdAt"`
}

// Utilize []CartItem in Order Struct
type CartItem struct {
	ArtworkID string  `firestore:"artworkId"`
	Quantity  int     `firestore:"quantity"`
	Price     float64 `firestore:"price"`
	// Print options for this item (can be extracted from artwork or set by user)
	PrintOptions PrintOrderOptions `firestore:"printOptions,omitempty"`
}

type Cart struct {
	BuyerID   string     `firestore:"buyerId"`
	Items     []CartItem `firestore:"items"`
	UpdatedAt time.Time  `firestore:"updatedAt"`
}

type Order struct {
	OrderID        string            `firestore:"orderId,omitempty"`
	BuyerID        string            `firestore:"buyerId"`
	PrintShopID    string            `firestore:"printShopId"`
	Items          []CartItem        `firestore:"items"`
	PrintOptions   PrintOrderOptions `firestore:"printOptions"` // Print configuration for the order
	TotalAmount    float64           `firestore:"totalAmount"`
	PaymentMethod  string            `firestore:"paymentMethod"`  // Legacy: "unpaid", "paid"
	TransactionID  string            `firestore:"transactionId"`  // Legacy: kept for backward compatibility
	PaymentStatus  string            `firestore:"paymentStatus"`  // "unpaid", "partial", "paid"
	PaymentID      string            `firestore:"paymentId"`      // Latest payment ID
	DeliveryStatus string            `firestore:"deliveryStatus"` // "pending", "processing", "ready", "delivered"
	DeliveryMethod string            `firestore:"deliveryMethod"` // "pickup", "shipping"
	PickupLocation string            `firestore:"pickupLocation"` // For pickup orders
	Status         string            `firestore:"status"`         // "pending", "confirmed", "processing", "ready", "completed"
	CreatedAt      time.Time         `firestore:"createdAt"`
	UpdatedAt      time.Time         `firestore:"updatedAt"`
}
