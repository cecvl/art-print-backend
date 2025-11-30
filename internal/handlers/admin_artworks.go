package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
)

// GetAdminArtworksHandler lists artworks filtered by processingStatus (query `status`)
func GetAdminArtworksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Filters: status, artistId, date range (from/to RFC3339)
	status := r.URL.Query().Get("status")
	artist := r.URL.Query().Get("artistId")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	q := firebase.FirestoreClient.Collection("artworks").OrderBy("createdAt", firestore.Desc)
	if status != "" {
		q = q.Where("processingStatus", "==", status)
	}
	if artist != "" {
		q = q.Where("artistId", "==", artist)
	}
	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			q = q.Where("createdAt", ">=", t)
		}
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			q = q.Where("createdAt", "<=", t)
		}
	}

	// Pagination: allow `limit` query param (default 100, max 500)
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > 500 {
				v = 500
			}
			limit = v
		}
	}

	snaps, err := q.Limit(limit).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("❌ Failed to fetch admin artworks: %v", err)
		http.Error(w, "Failed to fetch artworks", http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for _, s := range snaps {
		d := s.Data()
		d["id"] = s.Ref.ID
		results = append(results, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetAdminArtworkHandler returns full artwork doc for review (query `id`)
func GetAdminArtworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	doc, err := firebase.FirestoreClient.Collection("artworks").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("❌ Failed to get artwork %s: %v", id, err)
		http.Error(w, "Failed to fetch artwork", http.StatusInternalServerError)
		return
	}
	d := doc.Data()
	d["id"] = doc.Ref.ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// ResolveArtworkHandler allows admin to approve/reject/reprocess an artwork
// POST body JSON: { "id": "<artworkId>", "action": "approve"|"reject"|"reprocess", "note": "optional" }
func ResolveArtworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		ID     string `json:"id"`
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ID == "" || body.Action == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	artRef := firebase.FirestoreClient.Collection("artworks").Doc(body.ID)
	switch body.Action {
	case "approve":
		_, err := artRef.Set(ctx, map[string]interface{}{"processingStatus": "ready", "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
		if err != nil {
			log.Printf("❌ Failed to approve artwork %s: %v", body.ID, err)
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
	case "reject":
		_, err := artRef.Set(ctx, map[string]interface{}{"processingStatus": "failed", "processingErrors": []string{"rejected_by_admin"}, "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
		if err != nil {
			log.Printf("❌ Failed to reject artwork %s: %v", body.ID, err)
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
	case "reprocess":
		// reset status and enqueue job
		_, err := artRef.Set(ctx, map[string]interface{}{"processingStatus": "pending", "processingErrors": []string{}, "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
		if err != nil {
			log.Printf("❌ Failed to mark artwork %s for reprocess: %v", body.ID, err)
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
		// enqueue
		artDoc, _ := artRef.Get(ctx)
		cloudInfo := map[string]interface{}{"secureUrl": artDoc.Data()["imageUrl"], "publicId": artDoc.Data()["cloudinaryPublicId"], "folder": artDoc.Data()["cloudinaryFolder"]}
		_, _, _ = firebase.FirestoreClient.Collection("processing_queue").Add(ctx, map[string]interface{}{"artworkId": body.ID, "status": "pending", "createdAt": time.Now(), "cloudinary": cloudInfo})
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignArtworkHandler assigns an artwork to a print shop
// POST body JSON: { "id": "<artworkId>", "printShopId": "<shopId>" }
func AssignArtworkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		ID          string `json:"id"`
		PrintShopID string `json:"printShopId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ID == "" || body.PrintShopID == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	// set assignedTo on artwork and create assignment record
	assigned := map[string]interface{}{"printShopId": body.PrintShopID, "assignedAt": time.Now(), "assignedBy": ctx.Value("userId")}
	_, err := firebase.FirestoreClient.Collection("artworks").Doc(body.ID).Set(ctx, map[string]interface{}{"assignedTo": assigned}, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Failed to assign artwork %s: %v", body.ID, err)
		http.Error(w, "assignment failed", http.StatusInternalServerError)
		return
	}

	_, _, err = firebase.FirestoreClient.Collection("assignments").Add(ctx, map[string]interface{}{"artworkId": body.ID, "printShopId": body.PrintShopID, "status": "pending", "createdAt": time.Now(), "createdBy": ctx.Value("userId")})
	if err != nil {
		log.Printf("⚠️ Failed to create assignment record for artwork %s: %v", body.ID, err)
	}

	w.WriteHeader(http.StatusNoContent)
}
