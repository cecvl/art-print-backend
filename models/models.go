package models

import "time"

const (
	Buyer     = "buyer"
	Artist    = "artist"
	PrintShop = "printShop"
)

type User struct {
	Email     string    `firestore:"email"`
	UserType  string    `firestore:"userType"`
	CreatedAt time.Time `firestore:"createdAt"`
}

type Artwork struct {
	ID 			 string					`json:"id,omitempty"`
	ArtistID     string                 `firestore:"artistId"`
	Title        string                 `firestore:"title"`
	Description  string                 `firestore:"description"`
	ImageURL     string                 `firestore:"imageUrl"`
	PrintOptions map[string]interface{} `firestore:"printOptions"`
	IsAvailable  bool                   `firestore:"isAvailable"`
	CreatedAt    time.Time              `firestore:"createdAt"`
}

type Order struct {
	BuyerID        string            `firestore:"buyerId"`
	ArtworkID      string            `firestore:"artworkId"`
	PrintShopID    string            `firestore:"printShopId"`
	SelectedOptions map[string]string `firestore:"selectedOptions"`
	Status         string            `firestore:"status"`
	CreatedAt      time.Time         `firestore:"createdAt"`
	UpdatedAt      time.Time         `firestore:"updatedAt"`
}

