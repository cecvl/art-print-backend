package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"example.com/cloudinary-proxy/firebase"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// === GET PROFILE ===
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	doc, err := firebase.FirestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		log.Printf("❌ Failed to retrieve profile for user %s: %v", userID, err)
		http.Error(w, "User profile not found", http.StatusNotFound)
		return
	}

	profile := doc.Data()
	profile["uid"] = doc.Ref.ID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
	log.Printf("✅ Sent profile for user %s", userID)
}

// === UPDATE PROFILE TEXT FIELDS ===
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("❌ Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{
		"name":          true,
		"description":   true,
		"dateOfBirth":   true,
		"avatarUrl":     true,
		"backgroundUrl": true,
	}

	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}

	if len(filtered) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	filtered["updatedAt"] = time.Now()

	log.Printf("🔄 Updating profile for user %s with fields: %+v", userID, filtered)

	_, err := firebase.FirestoreClient.Collection("users").Doc(userID).Set(ctx, filtered, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Failed to update Firestore: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Profile updated for user %s", userID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated"})
}

// === UPLOAD AVATAR & BACKGROUND IMAGES ===
func UploadProfileAssetsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 UploadProfileAssetsHandler triggered")
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("❌ Failed to parse multipart form: %v", err)
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	// Cloudinary client setup
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		log.Printf("❌ Cloudinary init error: %v", err)
		http.Error(w, "Cloudinary setup failed", http.StatusInternalServerError)
		return
	}

	updates := make(map[string]interface{})

	// === Upload avatar ===
	avatarFile, avatarHeader, err := r.FormFile("avatar")
	log.Println("📥 Checking avatar field...")
	if err != nil {
		log.Printf("⚠️ No avatar file uploaded: %v", err)
	} else {
		defer avatarFile.Close()
		log.Printf("📤 Avatar file received: %s", avatarHeader.Filename)

		res, err := cld.Upload.Upload(ctx, avatarFile, uploader.UploadParams{
			Folder: "users/" + userID + "/profile",
		})
		if err != nil {
			log.Printf("❌ Avatar upload failed: %v", err)
			http.Error(w, "Avatar upload failed", http.StatusInternalServerError)
			return
		}
		updates["avatarUrl"] = res.SecureURL
		log.Printf("✅ Avatar uploaded: %s", res.SecureURL)
	}

	// === Upload background ===
	bgFile, bgHeader, err := r.FormFile("background")
	log.Println("📥 Checking background field...")

	if err != nil {
		log.Printf("⚠️ No background file uploaded: %v", err)
	} else {
		defer bgFile.Close()
		log.Printf("📤 Background file received: %s", bgHeader.Filename)

		res, err := cld.Upload.Upload(ctx, bgFile, uploader.UploadParams{
			Folder: "users/" + userID + "/profile",
		})
		if err != nil {
			log.Printf("❌ Background upload failed: %v", err)
			http.Error(w, "Background upload failed", http.StatusInternalServerError)
			return
		}
		updates["backgroundUrl"] = res.SecureURL
		log.Printf("✅ Background uploaded: %s", res.SecureURL)
	}

	if len(updates) == 0 {
		log.Println("⚠️ No files uploaded, no updates to Firestore")
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// Timestamp
	updates["updatedAt"] = time.Now()
	log.Printf("📝 Firestore update: %+v", updates)

	// Write to Firestore
	_, err = firebase.FirestoreClient.Collection("users").Doc(userID).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Firestore update failed: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Profile update complete for user %s", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile images uploaded"})
}
