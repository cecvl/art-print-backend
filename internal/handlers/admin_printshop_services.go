package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/google/uuid"

    "github.com/cecvl/art-print-backend/internal/firebase"
    "github.com/cecvl/art-print-backend/internal/models"
    "github.com/cecvl/art-print-backend/internal/repositories"
)

type adminCreateServiceReq struct {
    ShopID string             `json:"shopId"
    `
    Service models.PrintService `json:"service"`
}

// AdminCreateServiceHandler allows admins to create a service for a shop
func AdminCreateServiceHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var body struct {
        ShopID  string            `json:"shopId"`
        Service models.PrintService `json:"service"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if body.ShopID == "" || body.Service.Name == "" {
        http.Error(w, "shopId and service.name required", http.StatusBadRequest)
        return
    }

    svc := body.Service
    if svc.ID == "" {
        svc.ID = uuid.NewString()
    }
    svc.ShopID = body.ShopID
    svc.CreatedAt = time.Now()
    svc.UpdatedAt = time.Now()
    svc.IsActive = true

    repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
    if err := repo.CreateService(ctx, &svc); err != nil {
        log.Printf("❌ failed to create service: %v", err)
        http.Error(w, "failed to create service", http.StatusInternalServerError)
        return
    }

    // record service change
    _, _, _ = firebase.FirestoreClient.Collection("services_changes").Add(ctx, map[string]interface{}{"serviceId": svc.ID, "shopId": svc.ShopID, "action": "created", "details": svc, "createdAt": time.Now(), "createdBy": r.Context().Value("userId")})

    writeAdminAction(ctx, r, "create_service", "service", svc.ID, svc)

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]interface{}{"serviceId": svc.ID})
}

type updateServiceStatusReq struct {
    ServiceID string `json:"serviceId"`
    IsActive  bool   `json:"isActive"`
    Reason    string `json:"reason,omitempty"`
}

// UpdateServiceStatusHandler toggles a service active state and records change
func UpdateServiceStatusHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var body updateServiceStatusReq
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if body.ServiceID == "" {
        http.Error(w, "serviceId required", http.StatusBadRequest)
        return
    }

    repo := repositories.NewPrintShopRepository(firebase.FirestoreClient)
    updates := map[string]interface{}{"isActive": body.IsActive, "updatedAt": time.Now()}
    if err := repo.UpdateService(ctx, body.ServiceID, updates); err != nil {
        log.Printf("❌ failed to update service status: %v", err)
        http.Error(w, "failed to update service", http.StatusInternalServerError)
        return
    }

    action := "service_disabled"
    if body.IsActive {
        action = "service_enabled"
    }
    // write services_changes record
    _, _, _ = firebase.FirestoreClient.Collection("services_changes").Add(ctx, map[string]interface{}{"serviceId": body.ServiceID, "action": action, "reason": body.Reason, "createdAt": time.Now(), "createdBy": r.Context().Value("userId")})
    writeAdminAction(ctx, r, action, "service", body.ServiceID, map[string]interface{}{"isActive": body.IsActive, "reason": body.Reason})

    w.WriteHeader(http.StatusNoContent)
}
