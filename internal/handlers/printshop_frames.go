package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/cecvl/art-print-backend/internal/firebase"
    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
    "cloud.google.com/go/firestore"
)

// UploadFrameHandler allows authenticated print shop owners to upload frame images
func UploadFrameHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    uid := ctx.Value("userId")
    if uid == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    shopID := uid.(string)

    // parse form
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "failed to parse form", http.StatusBadRequest)
        return
    }
    file, fh, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "file required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    name := r.FormValue("name")
    description := r.FormValue("description")

    cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
    if err != nil {
        log.Printf("❌ cloudinary init failed: %v", err)
        http.Error(w, "cloudinary init failed", http.StatusInternalServerError)
        return
    }

    folder := "folder-one/frames/" + shopID
    useFilename := true
    uniqueFilename := true
    uploadRes, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
        Folder:         folder,
        PublicID:       fh.Filename,
        UseFilename:    &useFilename,
        UniqueFilename: &uniqueFilename,
    })
    if err != nil {
        log.Printf("❌ frame upload failed: %v", err)
        http.Error(w, "upload failed", http.StatusInternalServerError)
        return
    }

    // persist frame doc
    frameData := map[string]interface{}{
        "shopId":            shopID,
        "name":              name,
        "description":       description,
        "imageUrl":          uploadRes.SecureURL,
        "cloudinaryPublicId": uploadRes.PublicID,
        "cloudinaryFolder":  folder,
        "processingStatus":  "pending",
        "processingErrors":  []string{},
        "createdAt":         time.Now(),
    }
    docRef, _, err := firebase.FirestoreClient.Collection("frames").Add(ctx, frameData)
    if err != nil {
        log.Printf("❌ failed to save frame doc: %v", err)
        http.Error(w, "save failed", http.StatusInternalServerError)
        return
    }

    // enqueue processing job
    queueDoc := map[string]interface{}{
        "frameId": docRef.ID,
        "status":  "pending",
        "createdAt": time.Now(),
        "cloudinary": map[string]interface{}{
            "secureUrl": uploadRes.SecureURL,
            "publicId":  uploadRes.PublicID,
            "folder":    folder,
        },
    }
    if _, _, err := firebase.FirestoreClient.Collection("processing_queue").Add(ctx, queueDoc); err != nil {
        log.Printf("⚠️ failed to enqueue frame processing job: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"frameId": docRef.ID, "url": uploadRes.SecureURL})
}

// GetFramesHandler lists frames for the authenticated print shop
func GetFramesHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    uid := ctx.Value("userId")
    if uid == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    shopID := uid.(string)

    snaps, err := firebase.FirestoreClient.Collection("frames").Where("shopId", "==", shopID).OrderBy("createdAt", firestore.Desc).Documents(ctx).GetAll()
    if err != nil {
        log.Printf("❌ failed to fetch frames: %v", err)
        http.Error(w, "failed", http.StatusInternalServerError)
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

// RemoveFrameHandler removes a frame (soft-delete could be added instead)
func RemoveFrameHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    uid := ctx.Value("userId")
    if uid == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    shopID := uid.(string)

    var payload struct{
        FrameID string `json:"frameId"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if payload.FrameID == "" {
        http.Error(w, "frameId required", http.StatusBadRequest)
        return
    }

    // verify ownership
    docRef := firebase.FirestoreClient.Collection("frames").Doc(payload.FrameID)
    doc, err := docRef.Get(ctx)
    if err != nil || !doc.Exists() {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    if doc.Data()["shopId"] != shopID {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    if _, err := docRef.Delete(ctx); err != nil {
        log.Printf("❌ failed to delete frame %s: %v", payload.FrameID, err)
        http.Error(w, "delete failed", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
