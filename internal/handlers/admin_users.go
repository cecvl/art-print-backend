package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/cecvl/art-print-backend/internal/firebase"
)

// GetAdminUsersHandler lists users for admin console.
// Optional query params: role (array-contains), limit (int)
func GetAdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := r.URL.Query().Get("role")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	var iter *firestore.DocumentIterator
	if role != "" {
		iter = firebase.FirestoreClient.Collection("users").Where("roles", "array-contains", role).OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
	} else {
		iter = firebase.FirestoreClient.Collection("users").OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
	}

	docs, err := iter.GetAll()
	if err != nil {
		http.Error(w, "failed to query users", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		data := d.Data()
		// select fields to return
		userMap := map[string]interface{}{
			"uid":         data["uid"],
			"email":       data["email"],
			"name":        data["name"],
			"roles":       data["roles"],
			"description": data["description"],
			"avatarUrl":   data["avatarUrl"],
			"createdAt":   data["createdAt"],
		}
		out = append(out, userMap)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetAdminUserHandler returns a single user by uid (query param uid)
func GetAdminUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "missing uid", http.StatusBadRequest)
		return
	}

	doc, err := firebase.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(doc.Data())
}

type updateRolesRequest struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
}

// UpdateUserRolesHandler updates a user's roles and writes an admin action audit record
func UpdateUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body updateRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.UID == "" {
		http.Error(w, "missing uid", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"roles": body.Roles}
	if _, err := firebase.FirestoreClient.Collection("users").Doc(body.UID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to update roles", http.StatusInternalServerError)
		return
	}

	// write audit
	writeAdminAction(ctx, r, "update_roles", "user", body.UID, map[string]interface{}{"roles": body.Roles})

	w.WriteHeader(http.StatusNoContent)
}

type idRequest struct {
	UID string `json:"uid"`
}

// DeactivateUserHandler marks a user as inactive (sets isActive=false) and logs the admin action
func DeactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body idRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.UID == "" {
		http.Error(w, "missing uid", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"isActive": false}
	if _, err := firebase.FirestoreClient.Collection("users").Doc(body.UID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to deactivate user", http.StatusInternalServerError)
		return
	}

	writeAdminAction(ctx, r, "deactivate", "user", body.UID, nil)
	w.WriteHeader(http.StatusNoContent)
}

// ReactivateUserHandler sets isActive=true and logs the admin action
func ReactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body idRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.UID == "" {
		http.Error(w, "missing uid", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{"isActive": true}
	if _, err := firebase.FirestoreClient.Collection("users").Doc(body.UID).Set(ctx, updates, firestore.MergeAll); err != nil {
		http.Error(w, "failed to reactivate user", http.StatusInternalServerError)
		return
	}

	writeAdminAction(ctx, r, "reactivate", "user", body.UID, nil)
	w.WriteHeader(http.StatusNoContent)
}

// admin audit helper is in admin_audit.go
