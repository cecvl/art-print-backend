package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/cloudinary-proxy/cache"
	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/models"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadArtHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	// Verify user is an artist
	userDoc, err := firebase.FirestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		http.Error(w, "User verification failed", http.StatusInternalServerError)
		return
	}

	// Check if user has artist role
	var user models.User
	if err := userDoc.DataTo(&user); err != nil {
		log.Printf("Failed to parse user data: %v", err)
		http.Error(w, "User data parsing failed", http.StatusInternalServerError)
		return
	}

	hasArtistRole := false
	for _, role := range user.Roles {
		if role == models.Artist {
			hasArtistRole = true
			break
		}
	}

	if !hasArtistRole {
		http.Error(w, "Only artists can upload artworks", http.StatusForbidden)
		return
	}

	// Parse the multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get title and description from form
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Artwork title is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "Artwork description is required", http.StatusBadRequest)
		return
	}

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"), 
		os.Getenv("CLOUDINARY_API_KEY"), 
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		log.Printf("Cloudinary setup failed: %v", err)
		http.Error(w, "Cloudinary setup failed", http.StatusInternalServerError)
		return
	}

	// Upload to Cloudinary
	log.Printf("Uploading file: %s, size: %d bytes", fileHeader.Filename, fileHeader.Size)
	useFilename := true
	uniqueFilename := true

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "artworks/" + userID,
		PublicID:       fileHeader.Filename,
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
	})
	if err != nil {
		log.Printf("Cloudinary upload error: %v", err)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	// Create artwork document
	artwork := models.Artwork{
		ArtistID:     userID,
		Title:        title,
		Description:  description,
		ImageURL:     uploadResult.SecureURL,
		PrintOptions: make(map[string]interface{}), // Initialize empty print options
		IsAvailable:  true,
		CreatedAt:    time.Now(),
	}

	// Save to Firestore
	docRef, _, err := firebase.FirestoreClient.Collection("artworks").Add(ctx, artwork)
	if err != nil {
		log.Printf("Failed to save artwork: %v", err)
		http.Error(w, "Saving artwork failed", http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	cacheKey := "all_artworks"
	if err := cache.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("Failed to invalidate cache: %v", err)
	}

	// Prepare response
	response := map[string]interface{}{
		"id":          docRef.ID,
		"title":       artwork.Title,
		"description": artwork.Description,
		"imageUrl":    artwork.ImageURL,
		"artistId":    artwork.ArtistID,
		"createdAt":   artwork.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	log.Printf("Upload successful: %s", uploadResult.SecureURL)
}