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
}

type Cart struct {
	BuyerID   string     `firestore:"buyerId"`
	Items     []CartItem `firestore:"items"`
	UpdatedAt time.Time  `firestore:"updatedAt"`
}

type Order struct {
	OrderID       string     `firestore:"orderId,omitempty"`
	BuyerID       string     `firestore:"buyerId"`
	PrintShopID   string     `firestore:"printShopId"`
	Items         []CartItem `firestore:"items"`
	TotalAmount   float64    `firestore:"totalAmount"`
	PaymentMethod string     `firestore:"paymentMethod"`
	TransactionID string     `firestore:"transactionId"`
	Status        string     `firestore:"status"`
	CreatedAt     time.Time  `firestore:"createdAt"`
	UpdatedAt     time.Time  `firestore:"updatedAt"`
}
