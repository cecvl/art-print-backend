package models

import "time"

const (
	Buyer     = "buyer"
	Artist    = "artist"
	PrintShop = "printShop"
)

type User struct {
	UID           string    `firestore:"uid"`
	Email         string    `firestore:"email"`
	Roles         []string  `firestore:"roles"` //  ["artist", "buyer", "printShop"]
	Name          string    `firestore:"name"`
	DateOfBirth   string    `firestore:"dateOfBirth"`
	Description   string    `firestore:"description"`
	AvatarURL     string    `firestore:"avatarUrl"`
	BackgroundURL string    `firestore:"backgroundUrl"`
	CreatedAt     time.Time `firestore:"createdAt"`
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

type CartItem struct {
	ArtworkID string  `firestore:"artworkId"`
	Quantity  int     `firestore:"quantity"`
	Price     float64 `firestore:"price"`
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
