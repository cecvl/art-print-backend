package handlers

import (
	"net/http"
	"time"

	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/models"
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

	// Parse the multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	description := r.FormValue("description")

	// ✅ Replace with env values in production
	cld, err := cloudinary.NewFromParams("CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET")
	if err != nil {
		http.Error(w, "Cloudinary setup failed", http.StatusInternalServerError)
		return
	}

	// 🔧 Use pointers to bools
	useFilename := true
	uniqueFilename := true

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "folder-one/artworks/" + userID,
		PublicID:       fileHeader.Filename,
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
	})
	if err != nil {
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

	w.WriteHeader(http.StatusCreated)
}
