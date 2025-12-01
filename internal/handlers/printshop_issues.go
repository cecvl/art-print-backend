package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "time"

    "cloud.google.com/go/firestore"
    "github.com/cecvl/art-print-backend/internal/firebase"
)

type reportIssueReq struct {
    OrderID string `json:"orderId"`
    Issue   string `json:"issue"`
    Level   string `json:"level"` // e.g., "warning", "critical"
}

// PrintShopReportIssueHandler allows a print shop (authenticated) to report fulfillment issues
func PrintShopReportIssueHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var body reportIssueReq
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if body.OrderID == "" || body.Issue == "" {
        http.Error(w, "orderId and issue required", http.StatusBadRequest)
        return
    }

    // record in printshop_issues collection
    payload := map[string]interface{}{"orderId": body.OrderID, "issue": body.Issue, "level": body.Level, "createdAt": time.Now(), "reportedBy": r.Context().Value("userId")}
    if _, _, err := firebase.FirestoreClient.Collection("printshop_issues").Add(ctx, payload); err != nil {
        log.Printf("❌ failed to record printshop issue: %v", err)
        http.Error(w, "failed to record issue", http.StatusInternalServerError)
        return
    }

    // append admin note to order (best-effort)
    note := map[string]interface{}{"note": "printshop_issue: " + body.Issue, "createdAt": time.Now(), "createdBy": r.Context().Value("userId"), "level": body.Level}
    _, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Update(ctx, []firestore.Update{{Path: "adminNotes", Value: firestore.ArrayUnion(note)}, {Path: "status", Value: "error"}})
    if err != nil {
        log.Printf("⚠️ failed to append issue note to order: %v", err)
    }

    w.WriteHeader(http.StatusNoContent)
}
