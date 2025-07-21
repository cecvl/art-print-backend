package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/cloudinary-proxy/firebase"
)

// === GET PROFILE ===
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 🔒 Auth check
	userID := ctx.Value("userId")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uid := userID.(string)

	// 🔍 Get user document
	doc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		log.Printf("❌ Failed to retrieve profile for user %s: %v", uid, err)
		http.Error(w, "User profile not found", http.StatusNotFound)
		return
	}

	// 📦 Send response
	profile := doc.Data()
	profile["uid"] = doc.Ref.ID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)

	log.Printf("✅ Sent profile for user %s", uid)
}
