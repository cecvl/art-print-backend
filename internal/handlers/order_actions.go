package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

type setArtworkPrintShopsReq struct {
	ArtworkID          string   `json:"artworkId"`
	EligiblePrintShops []string `json:"eligiblePrintShops"`
}

// SetArtworkPrintShopsHandler allows an artist to set which print shops can process their artwork
// Authorization: artist who owns the artwork
func SetArtworkPrintShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := ""
	if v := ctx.Value("userId"); v != nil {
		if s, ok := v.(string); ok {
			uid = s
		}
	}
	if uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body setArtworkPrintShopsReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ArtworkID == "" {
		http.Error(w, "artworkId required", http.StatusBadRequest)
		return
	}

	// fetch artwork
	doc, err := firebase.FirestoreClient.Collection("artworks").Doc(body.ArtworkID).Get(ctx)
	if err != nil {
		http.Error(w, "artwork not found", http.StatusNotFound)
		return
	}
	var artwork models.Artwork
	if err := doc.DataTo(&artwork); err != nil {
		http.Error(w, "invalid artwork data", http.StatusInternalServerError)
		return
	}

	// check authorization: must be the artist who owns the artwork
	if artwork.ArtistID != uid {
		http.Error(w, "forbidden: only artwork owner can set print shops", http.StatusForbidden)
		return
	}

	// validate print shops exist
	for _, shopID := range body.EligiblePrintShops {
		if shopID == "" {
			continue
		}
		_, err := firebase.FirestoreClient.Collection("printShops").Doc(shopID).Get(ctx)
		if err != nil {
			log.Printf("⚠️ print shop %s not found, skipping", shopID)
			http.Error(w, "invalid print shop ID: "+shopID, http.StatusBadRequest)
			return
		}
	}

	// set eligiblePrintShops
	_, err = firebase.FirestoreClient.Collection("artworks").Doc(body.ArtworkID).Update(ctx, []firestore.Update{
		{Path: "eligiblePrintShops", Value: body.EligiblePrintShops},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		log.Printf("❌ failed to set eligible print shops for artwork: %v", err)
		http.Error(w, "failed to set print shops", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Artist %s set eligible print shops for artwork %s: %v", uid, body.ArtworkID, body.EligiblePrintShops)

	w.WriteHeader(http.StatusNoContent)
}
