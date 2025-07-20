package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/cloudinary-proxy/firebase"
)

type SessionLoginRequest struct {
	Token string `json:"token"`
}

func SessionLoginHandler(w http.ResponseWriter, r *http.Request) {
	var body SessionLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Token == "" {
		http.Error(w, "Missing ID token", http.StatusBadRequest)
		return
	}

	// Set session expiration (5 days)
	expiresIn := time.Hour * 24 * 5

	// Create the session cookie
	sessionCookie, err := firebase.AuthClient.SessionCookie(r.Context(), body.Token, expiresIn)
	if err != nil {
		log.Printf("Failed to create session cookie: %v", err)
		http.Error(w, "Failed to create session", http.StatusUnauthorized)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionCookie,
		MaxAge:   int(expiresIn.Seconds()),
		HttpOnly: true,
		Secure:   false, // set to true in prod with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}

func SessionLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie by setting it to expired
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true, // Use false only if testing over http
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
