package interfaces

type PrintSize struct {
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
	WidthCM    int     `json:"width"`
	HeightCM   int     `json:"height"`
}

type FrameOption struct {
	Type     string `json:"type"`
	Material string `json:"material"`
	Price    int    `json:"price"`
}

// MaterialOption represents the substrate (e.g., canvas, wood, paper, metal, linen)
type MaterialOption struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

// MediumOption represents the substance used (e.g., paint, ink, charcoal)
type MediumOption struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	BasePrice int    `json:"basePrice"`
}

type PrintOptionsResponse struct {
	Sizes     []PrintSize      `json:"sizes"`
	Frames    []FrameOption    `json:"frames"`
	Materials []MaterialOption `json:"materials"`
	Mediums   []MediumOption   `json:"mediums"`
}

// NEW INTERFACES FOR PRINT SHOP MANAGEMENT

type PrintServiceProvider interface {
	CalculatePrice(serviceID string, options PrintOrderOptions) (float64, error)
	ValidateOptions(options PrintOrderOptions) error
	GetAvailableServices(filters ServiceFilters) []PrintService
}

type PricingCalculator interface {
	CalculateBasePrice(service PrintService, size string, quantity int) float64
	ApplyMaterialModifier(basePrice float64, material string) float64
	ApplyMediumModifier(basePrice float64, medium string) float64
	ApplyQuantityDiscount(basePrice float64, quantity int, tiers []QuantityTier) float64
}

//type MatchingStrategy interface {
//FindPrintShops(order PrintOrder, availableShops []PrintShopProfile) []ShopMatch
//}

// Supporting types for interfaces
type PrintOrderOptions struct {
	Size      string `json:"size"`
	Material  string `json:"material"`
	Medium    string `json:"medium"`
	Frame     string `json:"frame"`
	Quantity  int    `json:"quantity"`
	RushOrder bool   `json:"rushOrder"`
}

type ServiceFilters struct {
	Technology string   `json:"technology"`
	Materials  []string `json:"materials"`
	MaxPrice   float64  `json:"maxPrice"`
	Location   string   `json:"location"`
}

type QuantityTier struct {
	MinQuantity int     `json:"minQuantity"`
	MaxQuantity int     `json:"maxQuantity"`
	Discount    float64 `json:"discount"`
}

type PrintService struct {
	ID          string      `json:"id"`
	ShopID      string      `json:"shopId"`
	Name        string      `json:"name"`
	
	// Technology describes the process (e.g., hand-made, machine, giclee). Optional.
	Technology  string      `json:"technology,omitempty"`
	
	// Variants offered by this service
	Materials   []MaterialOption `json:"materials"` // Substrates
	Mediums     []MediumOption   `json:"mediums"`   // Substances
	
	BasePrice   float64     `json:"basePrice"`
	PriceMatrix PriceMatrix `json:"priceMatrix"`
}

type PriceMatrix struct {
	SizeModifiers   map[string]float64 `json:"sizeModifiers"`
	QuantityTiers   []QuantityTier     `json:"quantityTiers"`
	MaterialMarkups map[string]float64 `json:"materialMarkups"`
	MediumMarkups   map[string]float64 `json:"mediumMarkups"`
}

type ShopMatch struct {
	ShopID       string  `json:"shopId"`
	ShopName     string  `json:"shopName"`
	ServiceID    string  `json:"serviceId"`
	TotalPrice   float64 `json:"totalPrice"`
	DeliveryDays int     `json:"deliveryDays"`
	MatchScore   float64 `json:"matchScore"`
}
