package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/handlers"
	"github.com/cecvl/art-print-backend/internal/middleware"
)

func main() {
	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // default to development
	}

	envPath := "configs/.env." + env

	// Load the environment file
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ No %s found, relying on exported env vars", envPath)
	} else {
		log.Printf("✅ Loaded environment from %s", envPath)
	}

	// Check Cloudinary secret presence
	secret := os.Getenv("CLOUDINARY_API_SECRET")
	if len(secret) == 0 {
		log.Println("⚠️ CLOUDINARY_API_SECRET not set!")
	} else {
		log.Printf("🔐 CLOUDINARY_API_SECRET length: %d", len(secret))
	}

	// Initialize Firebase
	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("🔥 Firebase initialization failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	mux := http.NewServeMux()

	// Routes
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	mux.Handle("/artworks/upload", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadArtHandler))))
	mux.Handle("/artworks", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtworksHandler)))
	mux.Handle("/getprofile", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetProfileHandler))))
	mux.Handle("/updateprofile", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler))))
	mux.Handle("/sessionLogin", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLoginHandler)))
	mux.Handle("/sessionLogout", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLogoutHandler)))
	mux.Handle("/cart/add", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.AddToCartHandler))))
	mux.Handle("/cart", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetCartHandler))))
	mux.Handle("/cart/remove", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.RemoveFromCartHandler))))
	mux.Handle("/checkout", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.CheckoutOrderHandler))))
	mux.Handle("/checkout/get", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetOrdersHandler))))

	// Apply CORS middleware
	handlerWithCORS := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Server running in %s mode on :%s", env, port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
