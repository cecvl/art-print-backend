package middleware

import (
	"context"
	"net/http"

	"example.com/cloudinary-proxy/firebase"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		token, err := firebase.AuthClient.VerifyIDToken(r.Context(), authHeader)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", token.UID)
		next(w, r.WithContext(ctx))
	}
}

