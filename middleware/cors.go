package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your local frontend + future Firebase Hosting
		allowedOrigins := map[string]bool{
			"http://localhost:3000": true,       // Your local frontend
			"https://your-firebase-app.web.app": true, // Future prod
		}

		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == "OPTIONS" {
            return // Preflight handled
		}

		next.ServeHTTP(w, r)
	})
}
