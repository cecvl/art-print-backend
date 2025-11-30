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

// GetAdminFramesHandler lists frames filtered by processingStatus, shopId, date range, and limit
func GetAdminFramesHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    status := r.URL.Query().Get("status")
    shop := r.URL.Query().Get("shopId")
    fromStr := r.URL.Query().Get("from")
    toStr := r.URL.Query().Get("to")

    q := firebase.FirestoreClient.Collection("frames").OrderBy("createdAt", firestore.Desc)
    if status != "" {
        q = q.Where("processingStatus", "==", status)
    }
    if shop != "" {
        q = q.Where("shopId", "==", shop)
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
        log.Printf("❌ Failed to fetch admin frames: %v", err)
        http.Error(w, "Failed to fetch frames", http.StatusInternalServerError)
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

// ResolveFrameHandler allows admin to approve/reject/reprocess a frame
// POST body JSON: { "id": "<frameId>", "action": "approve"|"reject"|"reprocess", "note": "optional" }
func ResolveFrameHandler(w http.ResponseWriter, r *http.Request) {
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

    frameRef := firebase.FirestoreClient.Collection("frames").Doc(body.ID)
    switch body.Action {
    case "approve":
        _, err := frameRef.Set(ctx, map[string]interface{}{"processingStatus": "ready", "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
        if err != nil {
            log.Printf("❌ Failed to approve frame %s: %v", body.ID, err)
            http.Error(w, "update failed", http.StatusInternalServerError)
            return
        }
    case "reject":
        _, err := frameRef.Set(ctx, map[string]interface{}{"processingStatus": "failed", "processingErrors": []string{"rejected_by_admin"}, "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
        if err != nil {
            log.Printf("❌ Failed to reject frame %s: %v", body.ID, err)
            http.Error(w, "update failed", http.StatusInternalServerError)
            return
        }
    case "reprocess":
        _, err := frameRef.Set(ctx, map[string]interface{}{"processingStatus": "pending", "processingErrors": []string{}, "admin": map[string]interface{}{"resolvedBy": ctx.Value("userId"), "resolvedAt": time.Now(), "resolutionNote": body.Note}}, firestore.MergeAll)
        if err != nil {
            log.Printf("❌ Failed to mark frame %s for reprocess: %v", body.ID, err)
            http.Error(w, "update failed", http.StatusInternalServerError)
            return
        }
        // enqueue job
        frameDoc, _ := frameRef.Get(ctx)
        cloudInfo := map[string]interface{}{"secureUrl": frameDoc.Data()["imageUrl"], "publicId": frameDoc.Data()["cloudinaryPublicId"], "folder": frameDoc.Data()["cloudinaryFolder"]}
        _, _, _ = firebase.FirestoreClient.Collection("processing_queue").Add(ctx, map[string]interface{}{"frameId": body.ID, "status": "pending", "createdAt": time.Now(), "cloudinary": cloudInfo})
    default:
        http.Error(w, "unknown action", http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
