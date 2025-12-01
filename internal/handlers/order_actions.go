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

type selectPrintShopReq struct {
    OrderID    string `json:"orderId"`
    PrintShopID string `json:"printShopId"`
}

// SelectPrintShopHandler allows an authorized user to set the print shop for an order
// Authorization: buyer who created the order OR an artist who owns any artwork in the order
func SelectPrintShopHandler(w http.ResponseWriter, r *http.Request) {
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

    var body selectPrintShopReq
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if body.OrderID == "" || body.PrintShopID == "" {
        http.Error(w, "orderId and printShopId required", http.StatusBadRequest)
        return
    }

    // fetch order
    doc, err := firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Get(ctx)
    if err != nil {
        http.Error(w, "order not found", http.StatusNotFound)
        return
    }
    var order models.Order
    if err := doc.DataTo(&order); err != nil {
        http.Error(w, "invalid order data", http.StatusInternalServerError)
        return
    }

    // only allow change if order is pending/confirmed (not completed)
    if order.Status == "completed" || order.Status == "cancelled" {
        http.Error(w, "order cannot be reassigned", http.StatusBadRequest)
        return
    }

    // check authorization: buyer OR artist owning any artwork in items
    authorized := false
    if order.BuyerID == uid {
        authorized = true
    } else {
        // check artwork ownership
        for _, it := range order.Items {
            artDoc, err := firebase.FirestoreClient.Collection("artworks").Doc(it.ArtworkID).Get(ctx)
            if err != nil {
                continue
            }
            var art models.Artwork
            if artDoc.DataTo(&art) == nil {
                if art.ArtistID == uid {
                    authorized = true
                    break
                }
            }
        }
    }
    if !authorized {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // set printShopId
    _, err = firebase.FirestoreClient.Collection("orders").Doc(body.OrderID).Update(ctx, []firestore.Update{{Path: "printShopId", Value: body.PrintShopID}, {Path: "updatedAt", Value: time.Now()}})
    if err != nil {
        log.Printf("❌ failed to set printshop for order: %v", err)
        http.Error(w, "failed to set printshop", http.StatusInternalServerError)
        return
    }

    writeAdminAction(ctx, r, "select_printshop", "order", body.OrderID, map[string]interface{}{"printShopId": body.PrintShopID})

    w.WriteHeader(http.StatusNoContent)
}
