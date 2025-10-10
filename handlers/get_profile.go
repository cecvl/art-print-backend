package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/firebase"
	"google.golang.org/api/iterator"
)


func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
    defer cancel()

    // Auth check
    userID := ctx.Value("userId")
    if userID == nil {
        http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
        return
    }
    uid := userID.(string)
    log.Printf("🔍 Fetching profile for user: %s", uid)

    // Get user document
    userDoc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
    if err != nil {
        log.Printf("❌ User fetch error: %v", err)
        http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
        return
    }

    if !userDoc.Exists() {
        log.Printf("❌ User document doesn't exist for UID: %s", uid)
        http.Error(w, `{"error": "User document not found"}`, http.StatusNotFound)
        return
    }

    userData := userDoc.Data()
    log.Printf("📋 User data: %+v", userData) // Debug log

    // Get artworks
    artworks := make([]map[string]interface{}, 0)
    iter := firebase.FirestoreClient.Collection("artworks").
        Where("artistId", "==", uid).
        OrderBy("createdAt", firestore.Desc).
        Limit(6).
        Documents(ctx)

    defer iter.Stop()

    artworkCount := 0
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Printf("⚠️ Artwork document error: %v", err)
            break
        }
        artworkCount++
        artworkData := doc.Data()
        //log.Printf("🎨 Artwork %d: %+v", artworkCount, artworkData) // Debug log
        artworkData["id"] = doc.Ref.ID
        artworks = append(artworks, artworkData)
    }

    log.Printf("✅ Found %d artworks for user %s", artworkCount, uid)

    response := map[string]interface{}{
        "user":     userData,
        "artworks": artworks,
    }
    response["user"].(map[string]interface{})["uid"] = uid

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("❌ Response encoding error: %v", err)
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
    }
}