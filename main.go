package main

import (
	"log"
	"net/http"
	"os"

	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/handlers"
	"example.com/cloudinary-proxy/middleware"
)

func main() {
	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("Firebase initialization failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	mux := http.NewServeMux()

	// Register your handlers
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	mux.Handle("signin", middleware.LogMiddleware(http.HandlerFunc(handlers.SignInHandler)))
	mux.Handle("/artworks", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadArtHandler))))
	mux.Handle("/orders", middleware.LogMiddleware(middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateOrderHandler))))

	// Wrap everything with CORS middleware
	handlerWithCORS := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
