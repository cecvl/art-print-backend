package middleware

import (
	"log"
	"net/http"
)

// LogMiddleware logs incoming HTTP requests.
func LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("➡️ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}
