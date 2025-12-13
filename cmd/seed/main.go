package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/seeders"
)

// loadEnv loads the seed environment
func loadEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// Map "production" to "prod" for filename
	filename := env
	if env == "production" {
		filename = "prod"
	}

	envPath := "configs/.env." + filename

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ No %s found, using system env vars", envPath)
	} else {
		log.Printf("🔧 Loaded environment from %s", envPath)
	}
	return env
}

func main() {
	env := loadEnv()

	if env != "dev" {
		log.Println("⛔ Seeders only run in dev mode")
		return
	}

	log.Println("🔥 Initializing Firebase...")
	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("❌ Firebase init failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	ctx := context.Background()

	log.Println("🌱 Running all seeders...")

	if err := seeders.SeedUsers(ctx, firebase.AuthClient, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Users seeder error: %v", err)
	}
	if err := seeders.SeedArtworks(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Artworks seeder error: %v", err)
	}
	if err := seeders.SeedCarts(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Carts seeder error: %v", err)
	}
	if err := seeders.SeedOrders(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Orders seeder error: %v", err)
	}
	if err := seeders.SeedPrintShops(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Print shops seeder error: %v", err)
	}
	if err := seeders.SeedPrintOptions(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Print options error: %v", err)
	}
	if err := seeders.SeedPricing(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Pricing error: %v", err)
	}

	log.Println("🎉 Seeding complete!")
}
