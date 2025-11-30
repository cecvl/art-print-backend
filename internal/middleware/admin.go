package middleware

import (
	"net/http"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/models"
)

// AdminOnly checks the authenticated user has the Admin role.
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userId")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		uid := userID.(string)

		// read user doc
		doc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(r.Context())
		if err != nil {
			http.Error(w, "Failed to verify user roles", http.StatusInternalServerError)
			return
		}
		rolesIfc := doc.Data()["roles"]
		if rolesIfc == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		roles, ok := rolesIfc.([]interface{})
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		for _, ritem := range roles {
			if s, ok := ritem.(string); ok && s == models.Admin {
				next(w, r)
				return
			}
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}
