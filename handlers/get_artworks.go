package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/cloudinary-proxy/cache"
	"example.com/cloudinary-proxy/firebase"
	"example.com/cloudinary-proxy/models"
)

const (
	cacheKey = "all_artworks"
)

func GetArtworksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cacheExpiration := 5 * time.Minute // Cache for 5 minutes

	// Try to get from cache first
	cachedArtworks, err := cache.GetCachedArtworks(ctx, cacheKey)
	if err == nil && len(cachedArtworks) > 0 {
		log.Println("✅ Serving artworks from cache")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cachedArtworks); err != nil {
			log.Printf("❌ Failed to encode cached artworks: %v", err)
			http.Error(w, "Encoding error", http.StatusInternalServerError)
		}
		return
	}

	// Cache miss or empty, fetch from database
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
			art.ID = doc.Ref.ID
			artworks = append(artworks, art)
		}
	}

	// Cache the results
	if len(artworks) > 0 {
		if err := cache.CacheArtworks(ctx, cacheKey, artworks, cacheExpiration); err != nil {
			log.Printf("⚠️ Failed to cache artworks: %v", err)
		} else {
			log.Println("✅ Artworks cached successfully")
		}
	}

	log.Printf("✅ Fetched %d artworks from database", len(artworks))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(artworks); err != nil {
		log.Printf("❌ Failed to encode artworks to JSON: %v", err)
		http.Error(w, "Encoding error", http.StatusInternalServerError)
	}
}