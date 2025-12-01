package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// GetAdminArtistsHandler lists artist users with admin-level fields
func GetAdminArtistsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docs, err := firebase.FirestoreClient.Collection("users").Where("roles", "array-contains", models.Artist).OrderBy("createdAt", firestore.Desc).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("❌ failed to query artists: %v", err)
		http.Error(w, "failed to query artists", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		data := d.Data()
		data["uid"] = d.Ref.ID
		out = append(out, data)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"artists": out})
}

// GetAdminArtistHandler returns full user doc for an artist (uid query)
func GetAdminArtistHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "uid required", http.StatusBadRequest)
		return
	}
	doc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	var u models.User
	if err := doc.DataTo(&u); err != nil {
		http.Error(w, "invalid user data", http.StatusInternalServerError)
		return
	}

	// include their artworks
	artDocs, _ := firebase.FirestoreClient.Collection("artworks").Where("artistId", "==", uid).OrderBy("createdAt", firestore.Desc).Documents(ctx).GetAll()
	var arts []models.Artwork
	for _, a := range artDocs {
		var ar models.Artwork
		if err := a.DataTo(&ar); err == nil {
			ar.ID = a.Ref.ID
			arts = append(arts, ar)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"artist": u, "artworks": arts})
}
