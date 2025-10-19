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

// InitFirebase initializes Firebase and connects either to emulators or production
func InitFirebase() error {
	ctx := context.Background()
	env := os.Getenv("APP_ENV") // either "development" or "production"

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "cloudinary-trial"
	}

	conf := &firebase.Config{
		ProjectID: projectID,
	}

	var app *firebase.App
	var err error

	// ---- DEVELOPMENT MODE ----
	if env == "dev" {
		// Set emulator environment variables (if not already set)
		if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
			os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
		}
		if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") == "" {
			os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")
		}

		log.Println("🔥 Running in DEVELOPMENT mode — connecting to Firebase Emulators")

		app, err = firebase.NewApp(ctx, conf)
		if err != nil {
			return fmt.Errorf("failed to init Firebase emulator app: %v", err)
		}

	} else {
		// ---- PRODUCTION MODE ----
		credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credFile == "" {
			credFile = "firebase-service-account.json"
		}

		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file not found: %s", credFile)
		}

		opt := option.WithCredentialsFile(credFile)
		app, err = firebase.NewApp(ctx, conf, opt)
		if err != nil {
			return fmt.Errorf("firebase init failed: %v", err)
		}

		log.Println("☁️ Running in PRODUCTION mode — connected to live Firebase")
	}

	// Initialize clients
	AuthClient, err = app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("auth init failed: %v", err)
	}

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("firestore init failed: %v", err)
	}

	log.Println("✅ Firebase initialized successfully")
	return nil
}
