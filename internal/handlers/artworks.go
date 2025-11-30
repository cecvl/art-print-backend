package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
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
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		http.Error(w, "Cloudinary setup failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Uploading file; %s, size: %d bytes", fileHeader.Filename, fileHeader.Size)
	// 🔧 Use pointers to boolsArts
	useFilename := true
	uniqueFilename := true

	// store originals under a clear path so worker can write preprocessed derivatives
	originalFolder := "folder-one/artworks/" + userID + "/original"
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         originalFolder,
		PublicID:       fileHeader.Filename,
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
	})
	if err != nil {
		log.Printf("Cloudianary upload error: %v", err)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	// Persist artwork document with processing status = pending
	artData := map[string]interface{}{
		"title":              title,
		"description":        description,
		"artistId":           userID,
		"imageUrl":           uploadResult.SecureURL,
		"cloudinaryPublicId": uploadResult.PublicID,
		"cloudinaryFolder":   originalFolder,
		"isAvailable":        true,
		"processingStatus":   "pending",
		"processingErrors":   []string{},
		"createdAt":          time.Now(),
	}

	docRef, _, err := firebase.FirestoreClient.Collection("artworks").Add(ctx, artData)
	if err != nil {
		log.Printf("❌ Saving artwork failed: %v", err)
		http.Error(w, "Saving artwork failed", http.StatusInternalServerError)
		return
	}

	// Enqueue a processing job in Firestore queue collection (simple queue)
	queueDoc := map[string]interface{}{
		"artworkId": docRef.ID,
		"status":    "pending",
		"createdAt": time.Now(),
		"cloudinary": map[string]interface{}{
			"secureUrl": uploadResult.SecureURL,
			"publicId":  uploadResult.PublicID,
			"folder":    originalFolder,
		},
	}
	if _, _, err := firebase.FirestoreClient.Collection("processing_queue").Add(ctx, queueDoc); err != nil {
		log.Printf("⚠️ Failed to enqueue processing job for artwork %s: %v", docRef.ID, err)
		// do not fail the upload — processing can be retried by a worker scanning artworks with pending status
	} else {
		log.Printf("✅ Enqueued processing job for artwork %s", docRef.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"url":              uploadResult.SecureURL,
		"artworkId":        docRef.ID,
		"processingStatus": "pending",
	})

	log.Printf("Upload successful: %s (artworkId=%s)", uploadResult.SecureURL, docRef.ID)
}

// GetArtworkStatusHandler returns processing status and analysis for an artwork
func GetArtworkStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// accept artwork id via query param `id` or `artworkId`
	id := r.URL.Query().Get("id")
	if id == "" {
		id = r.URL.Query().Get("artworkId")
	}
	if id == "" {
		http.Error(w, "missing artwork id", http.StatusBadRequest)
		return
	}

	doc, err := firebase.FirestoreClient.Collection("artworks").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("❌ Failed to fetch artwork %s: %v", id, err)
		http.Error(w, "Failed to fetch artwork", http.StatusInternalServerError)
		return
	}

	data := doc.Data()
	// pick relevant fields to return
	resp := map[string]interface{}{}
	for _, k := range []string{"processingStatus", "processingErrors", "analysis", "printReadyVersions", "imageUrl", "createdAt"} {
		if v, ok := data[k]; ok {
			resp[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("❌ Failed to encode artwork status for %s: %v", id, err)
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}
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
