package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// === UPDATE PROFILE ===
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 UpdateProfileHandler triggered")
	ctx := r.Context()

	// 🔒 Auth check
	userID := ctx.Value("userId")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uid := userID.(string)

	// 📄 Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("❌ Multipart form parse error: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	// ✏️ Text fields to be updated
	for _, field := range []string{"name", "description", "dateOfBirth"} {
		if val := r.FormValue(field); val != "" {
			updates[field] = val
		}
	}

	// ☁️ Cloudinary setup
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

	// 🖼️ Avatar upload
	if avatarFile, avatarHeader, err := r.FormFile("avatar"); err == nil {
		defer avatarFile.Close()
		log.Printf("📤 Avatar: %s", avatarHeader.Filename)

		res, err := cld.Upload.Upload(ctx, avatarFile, uploader.UploadParams{
			Folder: "users/" + uid + "/profile",
		})
		if err != nil {
			log.Printf("❌ Avatar upload failed: %v", err)
			http.Error(w, "Avatar upload failed", http.StatusInternalServerError)
			return
		}
		updates["avatarUrl"] = res.SecureURL
		log.Printf("✅ Avatar uploaded to: %s", res.SecureURL)
	}

	// 🌄 Background upload
	if bgFile, bgHeader, err := r.FormFile("background"); err == nil {
		defer bgFile.Close()
		log.Printf("📤 Background: %s", bgHeader.Filename)

		res, err := cld.Upload.Upload(ctx, bgFile, uploader.UploadParams{
			Folder: "users/" + uid + "/profile",
		})
		if err != nil {
			log.Printf("❌ Background upload failed: %v", err)
			http.Error(w, "Background upload failed", http.StatusInternalServerError)
			return
		}
		updates["backgroundUrl"] = res.SecureURL
		log.Printf("✅ Background uploaded to: %s", res.SecureURL)
	}

	// 🧾 Check if any updates exist
	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// 🕒 Add timestamp
	updates["updatedAt"] = time.Now()

	// 📝 Save to Firestore
	_, err = firebase.FirestoreClient.Collection("users").Doc(uid).Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Firestore update error: %v", err)
		http.Error(w, "Profile update failed", http.StatusInternalServerError)
		return
	}

	// ✅ Success
	log.Printf("✅ Profile updated for user %s", uid)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated"})
}
