package seeders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

type Artwork struct {
	Title       string `json:"Title"`
	Description string `json:"Description"`
	ImageURL    string `json:"ImageURL"`
	ArtistID    string `json:"ArtistID"`
}

// SeedArtworks loads artworks.json and writes them to Firestore
func SeedArtworks(ctx context.Context, client *firestore.Client) error {
	data, err := os.ReadFile("internal/seeders/artworks.json")
	if err != nil {
		return fmt.Errorf("failed to read artworks.json: %w", err)
	}

	var artworks []Artwork
	if err := json.Unmarshal(data, &artworks); err != nil {
		return fmt.Errorf("failed to parse artworks.json: %w", err)
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
