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
// handlers/profile.go
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 UpdateProfileHandler triggered")
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("❌ Failed to parse multipart form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	// ===== Text fields =====
	fields := []string{"name", "description", "dateOfBirth"}
	for _, field := range fields {
		val := r.FormValue(field)
		if val != "" {
			updates[field] = val
		}
	}

	// ===== Cloudinary Setup =====
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

	// ===== Avatar Upload =====
	if avatarFile, avatarHeader, err := r.FormFile("avatar"); err == nil {
		defer avatarFile.Close()
		log.Printf("📤 Avatar received: %s", avatarHeader.Filename)

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

	// ===== Background Upload =====
	if bgFile, bgHeader, err := r.FormFile("background"); err == nil {
		defer bgFile.Close()
		log.Printf("📤 Background received: %s", bgHeader.Filename)

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
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	updates["updatedAt"] = time.Now()

	// Write to Firestore
	_, err = firebase.FirestoreClient.Collection("users").Doc(userID).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Firestore update failed: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Profile updated for user %s", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated"})
}
