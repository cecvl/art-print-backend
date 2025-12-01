package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
)

// writeAdminAction records admin actions to `admin_actions` collection.
func writeAdminAction(ctx context.Context, r *http.Request, action, resourceType, resourceID string, details interface{}) {
	performedBy := "unknown"
	if v := r.Context().Value("userId"); v != nil {
		if s, ok := v.(string); ok {
			performedBy = s
		}
	}
	payload := map[string]interface{}{
		"action":       action,
		"resourceType": resourceType,
		"resourceId":   resourceID,
		"performedBy":  performedBy,
		"createdAt":    time.Now(),
		"details":      details,
	}
	if _, _, err := firebase.FirestoreClient.Collection("admin_actions").Add(ctx, payload); err != nil {
		log.Printf("⚠️ failed to write admin action: %v", err)
	}
}
