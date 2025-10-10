package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cecvl/art-print-backend/firebase"
	"github.com/cecvl/art-print-backend/models"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadArtHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("userId").(string)

	userDoc, err := firebase.FirestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil || userDoc.Data()["userType"] != models.Artist {
		http.Error(w, "Only artists can upload artworks", http.StatusForbidden)
		return
	}

	// Parse the multipart form (max 10MB)Kuza
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	description := r.FormValue("description")

	//temporary check for empty api key
	log.Println("API secret length:", len(os.Getenv("CLOUDINARY_API_SECRET")))
	// ✅ Replace with env values in production
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"),)
	if err != nil {
		http.Error(w, "Cloudinary setup failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Uploading file; %s, size: %d bytes", fileHeader.Filename, fileHeader.Size)
	// 🔧 Use pointers to boolsArts
	useFilename := true
	uniqueFilename := true

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "folder-one/artworks/" + userID,
		PublicID:       fileHeader.Filename,
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
	})
	if err != nil {
		log.Printf("Cloudianary upload error: %v", err)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	art := models.Artwork{
		Title:       title,
		Description: description,
		ImageURL:    uploadResult.SecureURL,
		ArtistID:    userID,
		CreatedAt:   time.Now(),
		IsAvailable: true,
	}

	_, _, err = firebase.FirestoreClient.Collection("artworks").Add(ctx, art)
	if err != nil {
		http.Error(w, "Saving artwork failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
    "url": uploadResult.SecureURL,
	})

	log.Printf("Upload successful: %s", uploadResult.SecureURL)
}

func GetArtworksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	snapshot, err := firebase.FirestoreClient.Collection("artworks").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("❌ Failed to fetch artworks: %v", err)
		http.Error(w, "Failed to fetch artworks", http.StatusInternalServerError)
		return
	}

	var artworks []models.Artwork
	for _, doc := range snapshot {
	var art models.Artwork
	if err := doc.DataTo(&art); err == nil {
		art.ID = doc.Ref.ID // Set the unique Firestore doc ID
		artworks = append(artworks, art)
	}
	}


	log.Printf("✅ Fetched %d artworks", len(artworks)) // 👈 Log confirmation

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(artworks); err != nil {
		log.Printf("❌ Failed to encode artworks to JSON: %v", err)
		http.Error(w, "Encoding error", http.StatusInternalServerError)
	}
}

