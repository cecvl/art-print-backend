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

type MaterialOption struct {
	Type       string  `json:"type"`
	Multiplier float64 `json:"multiplier"`
}

type MediumOption struct {
	Type      string `json:"type"`
	BasePrice int    `json:"basePrice"`
}

type PrintOptionsResponse struct {
	Sizes     []PrintSize      `json:"sizes"`
	Frames    []FrameOption    `json:"frames"`
	Materials []MaterialOption `json:"materials"`
	Mediums   []MediumOption   `json:"mediums"`
}
