package seeders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

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

// SeedUsers loads users.json, creates Auth accounts and Firestore profiles
func SeedUsers(ctx context.Context, authClient *auth.Client, fsClient *firestore.Client) error {
	data, err := os.ReadFile("internal/seeders/users.json")
	if err != nil {
		return fmt.Errorf("failed to read users.json: %w", err)
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to parse users.json: %w", err)
	}

	for _, u := range users {
		// Create user in Firebase Auth
		params := (&auth.UserToCreate{}).
			Email(u.Email).
			Password(u.Password).
			DisplayName(u.DisplayName)

		userRecord, err := authClient.CreateUser(ctx, params)
		if err != nil {
			log.Printf("⚠️ Could not create user %s: %v", u.Email, err)
			continue
		}

		// Create corresponding Firestore profile
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
