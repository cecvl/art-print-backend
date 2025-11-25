package handlers

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/cecvl/art-print-backend/internal/firebase"
    "github.com/cecvl/art-print-backend/internal/models"
)

// ArtistWithArtworks represents the response shape for an artist and their artworks
type ArtistWithArtworks struct {
    UID           string           `json:"uid"`
    Name          string           `json:"name"`
    Description   string           `json:"description,omitempty"`
    AvatarURL     string           `json:"avatarUrl,omitempty"`
    Artworks      []models.Artwork `json:"artworks"`
}

// GetArtistsHandler fetches users with role `artist` and includes their artworks
func GetArtistsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    log.Println("📥 GetArtistsHandler triggered")

    // Query users where roles array contains artist
    log.Printf("🔍 Fetching users with role '%s'...", models.Artist)
    userSnapshots, err := firebase.FirestoreClient.Collection("users").Where("roles", "array-contains", models.Artist).Documents(ctx).GetAll()
    if err != nil {
        log.Printf("❌ Failed to fetch artists: %v", err)
        http.Error(w, "Failed to fetch artists", http.StatusInternalServerError)
        return
    }
    log.Printf("✅ Fetched %d artists from Firestore", len(userSnapshots))

    var results []ArtistWithArtworks

    for _, uDoc := range userSnapshots {
        var u models.User
        if err := uDoc.DataTo(&u); err != nil {
            // skip malformed user doc but log it
            log.Printf("⚠️ Skipping user doc %s: %v", uDoc.Ref.ID, err)
            continue
        }

        uid := uDoc.Ref.ID

        // Fetch artworks for this artist
        artSnapshots, err := firebase.FirestoreClient.Collection("artworks").Where("artistId", "==", uid).Documents(ctx).GetAll()
        if err != nil {
            log.Printf("⚠️ Failed to fetch artworks for artist %s: %v", uid, err)
            // continue but return empty artworks list
        }

        var arts []models.Artwork
        for _, aDoc := range artSnapshots {
            var art models.Artwork
            if err := aDoc.DataTo(&art); err == nil {
                art.ID = aDoc.Ref.ID
                arts = append(arts, art)
            } else {
                log.Printf("⚠️ Skipping artwork doc %s: %v", aDoc.Ref.ID, err)
            }
        }

        results = append(results, ArtistWithArtworks{
            UID:           uid,
            Name:          u.Name,
            Description:   u.Description,
            AvatarURL:     u.AvatarURL,
            Artworks:      arts,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    log.Printf("📤 Encoding %d artists to JSON...", len(results))
    if err := json.NewEncoder(w).Encode(results); err != nil {
        log.Printf("❌ Failed to encode artists response: %v", err)
        http.Error(w, "Encoding error", http.StatusInternalServerError)
        return
    }
    log.Printf("✅ Successfully sent %d artists to client", len(results))
}
