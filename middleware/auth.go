package middleware

import (
	"context"
	"net/http"

	"github.com/cecvl/art-print-backend/firebase"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔐 Check for the 'session' cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Missing session cookie", http.StatusUnauthorized)
			return
		}

		// 🧾 Verify session cookie with Firebase Admin SDK
		token, err := firebase.AuthClient.VerifySessionCookie(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "Invalid session cookie", http.StatusUnauthorized)
			return
		}

		// ✅ Add UID to context for downstream handlers
		ctx := context.WithValue(r.Context(), "userId", token.UID)
		next(w, r.WithContext(ctx))
	}
}
