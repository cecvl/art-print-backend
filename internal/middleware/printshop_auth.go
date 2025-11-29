package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// PrintShopAuthMiddleware ensures the user is authenticated and is a print shop owner
func PrintShopAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// First check for authentication
		userID := r.Context().Value("userId")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		uid := userID.(string)

		// Get user document to check roles
		userDoc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(r.Context())
		if err != nil || !userDoc.Exists() {
			log.Printf("❌ User not found: %s", uid)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := userDoc.DataTo(&user); err != nil {
			log.Printf("❌ Failed to parse user data: %v", err)
			http.Error(w, "Invalid user data", http.StatusInternalServerError)
			return
		}

		// Check if user has printShop role
		hasPrintShopRole := false
		for _, role := range user.Roles {
			if role == models.PrintShop {
				hasPrintShopRole = true
				break
			}
		}

		if !hasPrintShopRole {
			log.Printf("❌ User %s does not have printShop role", uid)
			http.Error(w, "Access denied: Print shop owner required", http.StatusForbidden)
			return
		}

		// Add shop owner ID to context
		ctx := context.WithValue(r.Context(), "shopOwnerId", uid)
		next(w, r.WithContext(ctx))
	}
}

