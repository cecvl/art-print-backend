package firebase

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var (
	AuthClient      *auth.Client
	FirestoreClient *firestore.Client
)

// InitFirebase initializes Firebase app and its clients, supporting emulator mode
func InitFirebase() error {
	ctx := context.Background()

	// Detect emulator environment
	fsEmulator := os.Getenv("FIRESTORE_EMULATOR_HOST")
	authEmulator := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")

	var app *firebase.App
	var err error

	if fsEmulator != "" || authEmulator != "" {
		log.Println("🧩 Firebase Emulator detected")

		conf := &firebase.Config{
			ProjectID: os.Getenv("FIREBASE_PROJECT_ID"), // must be set even in emulator mode
		}

		app, err = firebase.NewApp(ctx, conf)
		if err != nil {
			return fmt.Errorf("failed to initialize Firebase app for emulator: %w", err)
		}
	} else {
		// Use real Firebase credentials
		credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credPath == "" {
			return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS not set")
		}

		opt := option.WithCredentialsFile(credPath)
		app, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			return fmt.Errorf("failed to create Firebase App: %w", err)
		}
		log.Println("🔐 Using production Firebase project")
	}

	// Initialize Auth client
	if AuthClient, err = app.Auth(ctx); err != nil {
		return fmt.Errorf("failed to init Auth: %w", err)
	}

	// Initialize Firestore client
	if FirestoreClient, err = app.Firestore(ctx); err != nil {
		return fmt.Errorf("failed to init Firestore: %w", err)
	}

	if fsEmulator != "" {
		log.Printf("🔥 Connected to Firestore Emulator at %s", fsEmulator)
	}
	if authEmulator != "" {
		log.Printf("👤 Connected to Auth Emulator at %s", authEmulator)
	}

	log.Println("✅ Firebase initialized successfully")
	return nil
}
