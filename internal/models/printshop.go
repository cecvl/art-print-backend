package models

import "time"

// PrintShopProfile represents a print service business
type PrintShopProfile struct {
	ID           string      `firestore:"id" json:"id"`
	OwnerID      string      `firestore:"ownerId" json:"ownerId"` // Links to User.UID
	Name         string      `firestore:"name" json:"name"`
	Description  string      `firestore:"description" json:"description"`
	Location     Location    `firestore:"location" json:"location"`
	Contact      ContactInfo `firestore:"contact" json:"contact"`
	Services     []string    `firestore:"services" json:"services"` // Service IDs
	IsActive     bool        `firestore:"isActive" json:"isActive"`
	Rating       float64     `firestore:"rating" json:"rating"`
	Capabilities []string    `firestore:"capabilities" json:"capabilities"`
	CreatedAt    time.Time   `firestore:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time   `firestore:"updatedAt" json:"updatedAt"`
}

// PrintService represents a specific service offered by a print shop
type PrintService struct {
	ID          string      `firestore:"id" json:"id"`
	ShopID      string      `firestore:"shopId" json:"shopId"`
	Name        string      `firestore:"name" json:"name"`
	Description string      `firestore:"description" json:"description"`
	Technology  string      `firestore:"technology" json:"technology"`
	BasePrice   float64     `firestore:"basePrice" json:"basePrice"`
	PriceMatrix PriceMatrix `firestore:"priceMatrix" json:"priceMatrix"`
	IsActive    bool        `firestore:"isActive" json:"isActive"`
	CreatedAt   time.Time   `firestore:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time   `firestore:"updatedAt" json:"updatedAt"`
}

// PriceMatrix for dynamic pricing
type PriceMatrix struct {
	SizeModifiers   map[string]float64 `firestore:"sizeModifiers" json:"sizeModifiers"`     // Size name -> multiplier
	QuantityTiers   []QuantityTier     `firestore:"quantityTiers" json:"quantityTiers"`     // Quantity discount tiers
	MaterialMarkups map[string]float64 `firestore:"materialMarkups" json:"materialMarkups"` // Material type -> markup multiplier
	MediumMarkups   map[string]float64 `firestore:"mediumMarkups" json:"mediumMarkups"`     // Medium type -> markup multiplier
	FramePrices     map[string]float64 `firestore:"framePrices" json:"framePrices"`         // Frame type -> price
	RushOrderFee    float64            `firestore:"rushOrderFee" json:"rushOrderFee"`       // Additional fee for rush orders
}

type QuantityTier struct {
	MinQuantity int     `firestore:"minQuantity" json:"minQuantity"`
	MaxQuantity int     `firestore:"maxQuantity" json:"maxQuantity"`
	Discount    float64 `firestore:"discount" json:"discount"`
}

// PrintOrderOptions for order matching
type PrintOrderOptions struct {
	Size      string `firestore:"size" json:"size"`
	Material  string `firestore:"material" json:"material"`
	Medium    string `firestore:"medium" json:"medium"`
	Frame     string `firestore:"frame" json:"frame"`
	Quantity  int    `firestore:"quantity" json:"quantity"`
	RushOrder bool   `firestore:"rushOrder" json:"rushOrder"`
}

// ShopMatch represents a matched print shop for an order
type ShopMatch struct {
	ShopID       string  `firestore:"shopId" json:"shopId"`
	ShopName     string  `firestore:"shopName" json:"shopName"`
	ServiceID    string  `firestore:"serviceId" json:"serviceId"`
	TotalPrice   float64 `firestore:"totalPrice" json:"totalPrice"`
	DeliveryDays int     `firestore:"deliveryDays" json:"deliveryDays"`
	MatchScore   float64 `firestore:"matchScore" json:"matchScore"`
	Technology   string  `firestore:"technology" json:"technology"`
}

// Supporting types
type Location struct {
	Address string `firestore:"address" json:"address"`
	City    string `firestore:"city" json:"city"`
	State   string `firestore:"state" json:"state"`
	Country string `firestore:"country" json:"country"`
}

type ContactInfo struct {
	Email   string `firestore:"email" json:"email"`
	Phone   string `firestore:"phone" json:"phone"`
	Website string `firestore:"website" json:"website"`
}

// Frame configuration for print shops
type Frame struct {
	ID          string    `firestore:"id" json:"id"`
	ShopID      string    `firestore:"shopId" json:"shopId"`
	Type        string    `firestore:"type" json:"type"`         // classic, modern, premium, minimalist
	Material    string    `firestore:"material" json:"material"` // wood, metal, acrylic, composite
	Name        string    `firestore:"name" json:"name"`         // Display name
	Description string    `firestore:"description" json:"description"`
	Price       float64   `firestore:"price" json:"price"` // Base price for this frame
	IsActive    bool      `firestore:"isActive" json:"isActive"`
	CreatedAt   time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// PrintSize configuration for print shops
type PrintSize struct {
	ID         string    `firestore:"id" json:"id"`
	ShopID     string    `firestore:"shopId" json:"shopId"`
	Name       string    `firestore:"name" json:"name"`             // A4, A3, 8x10, 11x14, 16x20, etc.
	WidthCM    float64   `firestore:"widthCM" json:"widthCM"`       // Width in centimeters
	HeightCM   float64   `firestore:"heightCM" json:"heightCM"`     // Height in centimeters
	WidthInch  float64   `firestore:"widthInch" json:"widthInch"`   // Width in inches (optional)
	HeightInch float64   `firestore:"heightInch" json:"heightInch"` // Height in inches (optional)
	Multiplier float64   `firestore:"multiplier" json:"multiplier"` // Price multiplier for this size
	IsActive   bool      `firestore:"isActive" json:"isActive"`
	CreatedAt  time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// Material/Substrate configuration for print shops
type Material struct {
	ID          string    `firestore:"id" json:"id"`
	ShopID      string    `firestore:"shopId" json:"shopId"`
	Type        string    `firestore:"type" json:"type"` // matte, glossy, canvas, metal, paper, premium
	Name        string    `firestore:"name" json:"name"` // Display name
	Description string    `firestore:"description" json:"description"`
	Markup      float64   `firestore:"markup" json:"markup"` // Multiplier for base price (e.g., 1.2 = 20% markup)
	IsActive    bool      `firestore:"isActive" json:"isActive"`
	CreatedAt   time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `firestore:"updatedAt" json:"updatedAt"`
}
