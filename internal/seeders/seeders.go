package seeders

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

//go:embed data/users.json
var usersData []byte

//go:embed data/artworks.json
var artworksData []byte

//go:embed data/carts.json
var cartsData []byte

//go:embed data/orders.json
var ordersData []byte

//go:embed data/printshops.json
var printshopsData []byte

//go:embed data/printoptions.json
var printoptionsData []byte

//go:embed data/pricing.json
var pricingData []byte

// Artwork model for seeding artworks
type Artwork struct {
	Title       string `json:"Title"`
	Description string `json:"Description"`
	ImageURL    string `json:"ImageURL"`
	ArtistID    string `json:"ArtistID"`
}

// User model for seeding users
type User struct {
	Email       string `json:"Email"`
	Password    string `json:"Password"`
	DisplayName string `json:"DisplayName"`
	Role        string `json:"Role"`
	Avatar      string `json:"Avatar"`
}

// CartItem and Cart models (match internal/models)
type CartItem struct {
	ArtworkID string  `json:"artworkId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	BuyerID   string     `json:"buyerId"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Order model for seeding orders (match internal/models)
type Order struct {
	OrderID       string     `json:"orderId"`
	BuyerID       string     `json:"buyerId"`
	PrintShopID   string     `json:"printShopId"`
	Items         []CartItem `json:"items"`
	TotalAmount   float64    `json:"totalAmount"`
	PaymentMethod string     `json:"paymentMethod"`
	TransactionID string     `json:"transactionId"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// Print Shop
type PrintShopSeed struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	ContactEmail string `json:"contactEmail"`
}

// Shop-specific print options
type PrintOptionSeed struct {
	PrintShopID string                   `json:"shopId"`
	Options     []map[string]interface{} `json:"options"`
}

// Shop-specific pricing
type PricingSeed struct {
	PrintShopID string                   `json:"shopId"`
	Pricing     []map[string]interface{} `json:"pricing"`
}

// SeedArtworks loads embedded artworks data and writes them to Firestore
func SeedArtworks(ctx context.Context, client *firestore.Client) error {
	var artworks []Artwork
	if err := json.Unmarshal(artworksData, &artworks); err != nil {
		return fmt.Errorf("failed to parse artworks data: %w", err)
	}

	for _, art := range artworks {
		_, _, err := client.Collection("artworks").Add(ctx, art)
		if err != nil {
			log.Printf("❌ Failed to seed artwork %s: %v", art.Title, err)
			continue
		}
		log.Printf("✅ Seeded artwork: %s", art.Title)
	}
	return nil
}

// SeedUsers loads embedded users data, creates Auth accounts and Firestore profiles
func SeedUsers(ctx context.Context, authClient *auth.Client, fsClient *firestore.Client) error {
	var users []User
	if err := json.Unmarshal(usersData, &users); err != nil {
		return fmt.Errorf("failed to parse users data: %w", err)
	}

	for _, u := range users {
		params := (&auth.UserToCreate{}).
			Email(u.Email).
			Password(u.Password).
			DisplayName(u.DisplayName)

		userRecord, err := authClient.CreateUser(ctx, params)
		if err != nil {
			log.Printf("⚠️ Could not create user %s: %v", u.Email, err)
			continue
		}

		profile := map[string]interface{}{
			"uid":         userRecord.UID,
			"displayName": u.DisplayName,
			"email":       u.Email,
			"role":        u.Role,
			"avatar":      u.Avatar,
			"createdAt":   firestore.ServerTimestamp,
		}

		_, err = fsClient.Collection("users").Doc(userRecord.UID).Set(ctx, profile)
		if err != nil {
			log.Printf("⚠️ Failed to create Firestore profile for %s: %v", u.Email, err)
			continue
		}

		log.Printf("✅ Seeded user: %s (%s)", u.DisplayName, u.Email)
	}
	return nil
}

// SeedCarts loads embedded carts data and writes them to Firestore
func SeedCarts(ctx context.Context, client *firestore.Client) error {
	var carts []Cart
	if err := json.Unmarshal(cartsData, &carts); err != nil {
		return fmt.Errorf("failed to parse carts data: %w", err)
	}

	for _, cart := range carts {
		_, err := client.Collection("carts").Doc(cart.BuyerID).Set(ctx, cart)
		if err != nil {
			log.Printf("❌ Failed to seed cart for buyer %s: %v", cart.BuyerID, err)
			continue
		}
		log.Printf("✅ Seeded cart for buyer: %s", cart.BuyerID)
	}
	return nil
}

// SeedOrders loads embedded orders data and writes them to Firestore
func SeedOrders(ctx context.Context, client *firestore.Client) error {
	var orders []Order
	if err := json.Unmarshal(ordersData, &orders); err != nil {
		return fmt.Errorf("failed to parse orders data: %w", err)
	}

	for _, order := range orders {
		_, err := client.Collection("orders").Doc(order.OrderID).Set(ctx, order)
		if err != nil {
			log.Printf("❌ Failed to seed order %s: %v", order.OrderID, err)
			continue
		}
		log.Printf("✅ Seeded order: %s", order.OrderID)
	}
	return nil
}

// SeedPrintShops loads embedded printshops data and writes them to Firestore
func SeedPrintShops(ctx context.Context, client *firestore.Client) error {
	var shops []PrintShopSeed
	if err := json.Unmarshal(printshopsData, &shops); err != nil {
		return fmt.Errorf("failed to parse printshops data: %w", err)
	}

	for _, shop := range shops {
		ref := client.Collection("printshops").Doc(shop.ID)

		if doc, _ := ref.Get(ctx); doc.Exists() {
			continue
		}

		_, err := ref.Set(ctx, shop)
		if err != nil {
			log.Printf("❌ Failed to seed print shop %s: %v", shop.ID, err)
			continue
		}
		log.Printf("✅ Seeded print shop: %s", shop.Name)
	}
	return nil
}

// SeedPrintOptions loads embedded printoptions data and writes them to Firestore
func SeedPrintOptions(ctx context.Context, client *firestore.Client) error {
	var entries []PrintOptionSeed
	if err := json.Unmarshal(printoptionsData, &entries); err != nil {
		return fmt.Errorf("failed to parse printoptions data: %w", err)
	}

	for _, entry := range entries {
		for _, opt := range entry.Options {
			optID := opt["id"].(string)

			ref := client.Collection("printshops").
				Doc(entry.PrintShopID).
				Collection("printoptions").
				Doc(optID)

			if doc, _ := ref.Get(ctx); doc.Exists() {
				continue
			}

			_, err := ref.Set(ctx, opt)
			if err != nil {
				log.Printf("❌ Failed to seed option %s for shop %s: %v", optID, entry.PrintShopID, err)
				continue
			}
			log.Printf("✅ Seeded print option: %s for shop %s", optID, entry.PrintShopID)
		}
	}
	return nil
}

// SeedPricing loads embedded pricing data and writes them to Firestore
func SeedPricing(ctx context.Context, client *firestore.Client) error {
	var entries []PricingSeed
	if err := json.Unmarshal(pricingData, &entries); err != nil {
		return fmt.Errorf("failed to parse pricing data: %w", err)
	}

	for _, entry := range entries {
		for _, price := range entry.Pricing {
			optionID := price["optionId"].(string)

			ref := client.Collection("printshops").
				Doc(entry.PrintShopID).
				Collection("pricing").
				Doc(optionID)

			if doc, _ := ref.Get(ctx); doc.Exists() {
				continue
			}

			_, err := ref.Set(ctx, price)
			if err != nil {
				log.Printf("❌ Failed to seed pricing %s for shop %s: %v", optionID, entry.PrintShopID, err)
				continue
			}
			log.Printf("✅ Seeded pricing: %s for shop %s", optionID, entry.PrintShopID)
		}
	}
	return nil
}
