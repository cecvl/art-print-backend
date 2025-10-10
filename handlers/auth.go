package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cecvl/art-print-backend/firebase"
	"github.com/cecvl/art-print-backend/models"
	"firebase.google.com/go/auth"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		UserType string `json:"userType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("Received signup request: email=%s, userType=%s", req.Email, req.UserType)

	validTypes := map[string]bool{models.Buyer: true, models.Artist: true, models.PrintShop: true}
	if !validTypes[req.UserType] {
		log.Printf("Invalid user type provided: %s", req.UserType)
		http.Error(w, "Invalid user type", http.StatusBadRequest)
		return
	}

	user, err := firebase.AuthClient.CreateUser(r.Context(), (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password),
	)
	if err != nil {
		log.Printf("Failed to create Firebase user: %v", err)
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}
	log.Printf("Created Firebase user: UID=%s", user.UID)

	_, err = firebase.FirestoreClient.Collection("users").Doc(user.UID).Set(r.Context(), models.User{
	Email:     req.Email,
	Roles:     []string{req.UserType}, // assign single role as array
	CreatedAt: time.Now(),
})

	if err != nil {
		log.Printf("Failed to write user to Firestore: %v. Rolling back Firebase user.", err)
		firebase.AuthClient.DeleteUser(r.Context(), user.UID)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	//Return Token
	customToken, err := firebase.AuthClient.CustomToken(r.Context(), user.UID)
	if err != nil {
		log.Printf("Failed to create custom token: %v", err)
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	log.Printf("User stored in Firestore and token generated: UID=%s", user.UID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": customToken})
}


