package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/handlers"
	"github.com/cecvl/art-print-backend/internal/middleware"
	"github.com/cecvl/art-print-backend/internal/seeders"
)

// loadEnv loads the environment variables based on APP_ENV
func loadEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	envPath := "configs/.env." + env
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ No %s found, relying on system env vars", envPath)
	} else {
		log.Printf("✅ Loaded environment from %s", envPath)
	}
	return env
}

// runSeeders seeds demo data in development mode
func runSeeders(env string) {
	if env != "dev" {
		return
	}

	ctx := context.Background()
	log.Println("🌱 Seeding Firestore and Auth with demo data...")

	if err := seeders.SeedUsers(ctx, firebase.AuthClient, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Seeder (users) error: %v", err)
	}
	if err := seeders.SeedArtworks(ctx, firebase.FirestoreClient); err != nil {
		log.Printf("⚠️ Seeder (artworks) error: %v", err)
	}
	// if you have carts/orders seeders, add them here:
	// if err := seeders.SeedCarts(...); err != nil { ... }
	// if err := seeders.SeedOrders(...); err != nil { ... }
}

// setupRoutes initializes all routes
func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	mux.Handle("/sessionLogin", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLoginHandler)))
	mux.Handle("/sessionLogout", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLogoutHandler)))
	mux.Handle("/artworks", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtworksHandler)))

	// Authenticated routes
	protected := middleware.AuthMiddleware
	mux.Handle("/artworks/upload", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.UploadArtHandler))))
	mux.Handle("/getprofile", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.GetProfileHandler))))
	mux.Handle("/updateprofile", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.UpdateProfileHandler))))
	mux.Handle("/cart/add", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.AddToCartHandler))))
	mux.Handle("/cart/remove", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.RemoveFromCartHandler))))
	mux.Handle("/cart", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.GetCartHandler))))
	mux.Handle("/checkout", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.CheckoutHandler))))
	mux.Handle("/orders", middleware.LogMiddleware(protected(http.HandlerFunc(handlers.GetOrdersHandler))))

	return middleware.CORS(mux)
}

func main() {
	env := loadEnv()

	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("🔥 Firebase initialization failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	runSeeders(env)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	handler := setupRoutes()
	log.Printf("🚀 Server running in %s mode on :%s", env, port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
