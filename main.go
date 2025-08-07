package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"example.com/cloudinary-proxy/cache"
	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/handlers"
	"example.com/cloudinary-proxy/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env FILE FOUND, relying on exported env vars")
	}

	//check if API keys are loaded

	log.Printf("CLOUDINARY_API_SECRET length: %d", len(os.Getenv("CLOUDINARY_API_SECRET")))

	//INitialize Redis
	cache.InitRedis()

	// Create a context with timeout for Redis connection test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test Redis connection
	_, err := cache.RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("Firebase initialization failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	mux := http.NewServeMux()

	// 🔐 Auth route
	// Login is implemented in sessions.go (SessionLoginHandler)
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	

	// 🖼️ Artworks
	mux.Handle("/artworks/upload", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadArtHandler))))
	mux.Handle("/artworks", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtworksHandler)))
	
	// 🧑🎨 Profile
	mux.Handle("/getprofile", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetProfileHandler))))
	mux.Handle("/updateprofile", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler))))
	
	// Sessions
	mux.Handle("/sessionLogin", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLoginHandler)))
	mux.Handle("/sessionLogout", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLogoutHandler)))
	
	// 🛒 Cart
	mux.Handle("/cart/add", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.AddToCartHandler))))
	mux.Handle("/cart", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetCartHandler))))
	mux.Handle("/cart/remove", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.RemoveFromCartHandler))))

	// 💳 Checkout
	mux.Handle("/checkout", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.CheckoutOrderHandler))))
	mux.Handle("checkout/get", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetOrdersHandler))))
	
	// 🌍 CORS
	handlerWithCORS := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
