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

	seeders.SeedUsers(ctx, firebase.AuthClient, firebase.FirestoreClient)
	seeders.SeedArtworks(ctx, firebase.FirestoreClient)
	seeders.SeedCarts(ctx, firebase.FirestoreClient)
	seeders.SeedOrders(ctx, firebase.FirestoreClient)
	seeders.SeedPrintShops(ctx, firebase.FirestoreClient)
	seeders.SeedPrintOptions(ctx, firebase.FirestoreClient)
	seeders.SeedPricing(ctx, firebase.FirestoreClient)
}

// setupRoutes initializes all routes
func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Print shop console handlers
	printOptionsHandler := handlers.NewPrintOptionsHandler()
	pricingHandler := handlers.NewPricingHandler()
	printShopConsoleHandler := handlers.NewPrintShopConsoleHandler()
	printShopConfigHandler := handlers.NewPrintShopConfigHandler()
	printShopServiceConfigHandler := handlers.NewPrintShopServiceConfigHandler()
	publicPrintShopHandler := handlers.NewPublicPrintShopHandler()
	matchingHandler := handlers.NewMatchingHandler()
	paymentHandler := handlers.NewPaymentHandler()

	// Health check route (no logging middleware for efficiency)
	mux.Handle("/health", http.HandlerFunc(handlers.HealthHandler))

	// Public routes
	mux.Handle("/signup", middleware.LogMiddleware(http.HandlerFunc(handlers.SignUpHandler)))
	mux.Handle("/sessionLogin", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLoginHandler)))
	mux.Handle("/sessionLogout", middleware.LogMiddleware(http.HandlerFunc(handlers.SessionLogoutHandler)))
	mux.Handle("/artworks", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtworksHandler)))
	mux.Handle("/artists", middleware.LogMiddleware(http.HandlerFunc(handlers.GetArtistsHandler)))

	// Print options route
	mux.Handle("/print-options", middleware.LogMiddleware(http.HandlerFunc(printOptionsHandler.GetPrintOptions)))

	// Public print shop endpoints (no authentication required)
	mux.Handle("/printshops", middleware.LogMiddleware(http.HandlerFunc(publicPrintShopHandler.GetActiveShops)))
	mux.Handle("/printshops/details", middleware.LogMiddleware(http.HandlerFunc(publicPrintShopHandler.GetShopDetails)))
	mux.Handle("/printshops/match", middleware.LogMiddleware(http.HandlerFunc(publicPrintShopHandler.MatchShopsForOrder)))
	mux.Handle("/printshops/calculate", middleware.LogMiddleware(http.HandlerFunc(publicPrintShopHandler.CalculatePriceForService)))

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
	//calculate price
	mux.Handle("/calculate-price", middleware.LogMiddleware(protected(http.HandlerFunc(pricingHandler.CalculatePrice))))

	// Print Shop Console routes (requires printShop role)
	// Chain: AuthMiddleware -> PrintShopAuthMiddleware -> Handler
	printShopChain := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.AuthMiddleware(middleware.PrintShopAuthMiddleware(h))
	}

	// Shop profile management
	mux.Handle("/printshop/profile", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.GetShopProfile)))
	mux.Handle("/printshop/profile/create", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.CreateShopProfile)))
	mux.Handle("/printshop/profile/update", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.UpdateShopProfile)))

	// Service management
	mux.Handle("/printshop/services", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.GetServices)))
	mux.Handle("/printshop/services/create", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.CreateService)))
	mux.Handle("/printshop/services/update/", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.UpdateService)))
	mux.Handle("/printshop/services/delete/", middleware.LogMiddleware(printShopChain(printShopConsoleHandler.DeleteService)))

	// Configuration management - Frames
	mux.Handle("/printshop/frames", middleware.LogMiddleware(printShopChain(printShopConfigHandler.GetFrames)))
	mux.Handle("/printshop/frames/create", middleware.LogMiddleware(printShopChain(printShopConfigHandler.CreateFrame)))
	mux.Handle("/printshop/frames/update/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.UpdateFrame)))
	mux.Handle("/printshop/frames/delete/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.DeleteFrame)))

	// Configuration management - Sizes
	mux.Handle("/printshop/sizes", middleware.LogMiddleware(printShopChain(printShopConfigHandler.GetSizes)))
	mux.Handle("/printshop/sizes/create", middleware.LogMiddleware(printShopChain(printShopConfigHandler.CreateSize)))
	mux.Handle("/printshop/sizes/update/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.UpdateSize)))
	mux.Handle("/printshop/sizes/delete/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.DeleteSize)))

	// Configuration management - Materials
	mux.Handle("/printshop/materials", middleware.LogMiddleware(printShopChain(printShopConfigHandler.GetMaterials)))
	mux.Handle("/printshop/materials/create", middleware.LogMiddleware(printShopChain(printShopConfigHandler.CreateMaterial)))
	mux.Handle("/printshop/materials/update/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.UpdateMaterial)))
	mux.Handle("/printshop/materials/delete/", middleware.LogMiddleware(printShopChain(printShopConfigHandler.DeleteMaterial)))

	// Service pricing configuration
	mux.Handle("/printshop/services/pricing/", middleware.LogMiddleware(printShopChain(printShopServiceConfigHandler.GetServicePricing)))
	mux.Handle("/printshop/services/pricing/update/", middleware.LogMiddleware(printShopChain(printShopServiceConfigHandler.UpdateServicePricing)))
	mux.Handle("/printshop/services/calculate/", middleware.LogMiddleware(printShopChain(printShopServiceConfigHandler.CalculateServicePrice)))

	// Order matching endpoints (admin/authenticated)
	mux.Handle("/orders/matches", middleware.LogMiddleware(protected(http.HandlerFunc(matchingHandler.GetOrderMatches))))
	mux.Handle("/orders/assign", middleware.LogMiddleware(protected(http.HandlerFunc(matchingHandler.AssignShopToOrder))))

	// Payment endpoints
	mux.Handle("/payments/create", middleware.LogMiddleware(protected(http.HandlerFunc(paymentHandler.CreatePaymentHandler))))
	mux.Handle("/payments/verify", middleware.LogMiddleware(protected(http.HandlerFunc(paymentHandler.VerifyPaymentHandler))))
	mux.Handle("/payments", middleware.LogMiddleware(protected(http.HandlerFunc(paymentHandler.GetPaymentsHandler))))
	mux.Handle("/payments/webhook/", middleware.LogMiddleware(http.HandlerFunc(paymentHandler.PaymentWebhookHandler)))
	mux.Handle("/payments/refund", middleware.LogMiddleware(protected(http.HandlerFunc(paymentHandler.RefundPaymentHandler))))

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
