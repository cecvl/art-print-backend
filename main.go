package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/handlers"
	"example.com/cloudinary-proxy/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env FILE FOUND, relying on exported env vars")
	}

	log.Printf("CLOUDINARY_API_SECRET length: %d", len(os.Getenv("CLOUDINARY_API_SECRET")))

	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("Firebase initialization failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	mux := http.NewServeMux()

	// 🔐 Auth routes
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	mux.Handle("/signin", middleware.LogMiddleware(http.HandlerFunc(handlers.SignInHandler)))

	// 🖼️ Artworks
	mux.Handle("/artworks/upload", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadArtHandler))))
	mux.Handle("/artworks", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtworksHandler)))

	// 🛒 Cart
	mux.Handle("/cart/add", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.AddToCartHandler))))
	mux.Handle("/cart", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetCartHandler))))
	mux.Handle("/cart/remove", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.RemoveFromCartHandler))))

	// 💳 Checkout
	mux.Handle("/checkout", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.CheckoutHandler))))

	// 🧑🎨 Profile
	mux.Handle("/profile", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.GetProfileHandler))))
	mux.Handle("/profile/update", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler))))
	mux.Handle("/profile/upload-assets", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadProfileAssetsHandler))))

	// 🛍️ Orders
	// 🧾 Fallback single-order handler (legacy)
	mux.Handle("/orders", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateOrderHandler))))

	// 🌍 CORS
	handlerWithCORS := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
