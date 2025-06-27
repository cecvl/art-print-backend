package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/cloudinary-proxy/firebase"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer "){
			http.Error(w, "Authorization header must start with Bearer", http.StatusUnauthorized)
			return
		}

		//Extract only the token
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := firebase.AuthClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", token.UID)
		next(w, r.WithContext(ctx))
	}
}

