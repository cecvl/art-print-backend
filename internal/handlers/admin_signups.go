package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// GetAdminSignupsHandler lists recent or pending signups for admin review
// Filters: role (optional), isActive=false (pending)
func GetAdminSignupsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := r.URL.Query().Get("role")

	q := firebase.FirestoreClient.Collection("users").Where("isActive", "==", false)
	if role != "" {
		q = q.Where("roles", "array-contains", role)
	}

	docs, err := q.OrderBy("createdAt", firestore.Desc).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("❌ failed to query signups: %v", err)
		http.Error(w, "failed to query signups", http.StatusInternalServerError)
		return
	}

	out := make([]*models.User, 0, len(docs))
	for _, d := range docs {
		var u models.User
		if err := d.DataTo(&u); err == nil {
			out = append(out, &u)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"signups": out})
}
