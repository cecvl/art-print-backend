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

// InitFirebase initializes Firebase based on the environment (dev or prod)
func InitFirebase() error {
	ctx := context.Background()
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod" // default fallback
	}

	cfg := buildFirebaseConfig(env)

	app, err := createFirebaseApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	// Initialize clients
	if AuthClient, err = app.Auth(ctx); err != nil {
		return fmt.Errorf("failed to init Firebase Auth: %w", err)
	}

	if FirestoreClient, err = app.Firestore(ctx); err != nil {
		return fmt.Errorf("failed to init Firestore: %w", err)
	}

	log.Println("✅ Firebase initialized successfully")
	return nil
}

// buildFirebaseConfig sets up Firebase config and environment variables
func buildFirebaseConfig(env string) *firebase.Config {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "cloudinary-trial"
	}

	if env == "dev" || env == "development" {
		setEmulatorVars()
		log.Println("🔥 Running in DEVELOPMENT mode — using Firebase Emulators")
	} else {
		log.Println("☁️ Running in PRODUCTION mode — connecting to live Firebase")
	}

	return &firebase.Config{
		ProjectID: projectID,
	}
}

// createFirebaseApp initializes Firebase app using credentials or emulators
func createFirebaseApp(ctx context.Context, cfg *firebase.Config) (*firebase.App, error) {
	env := os.Getenv("APP_ENV")

	if env == "dev" || env == "development" {
		return firebase.NewApp(ctx, cfg)
	}

	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credFile == "" {
		credFile = "firebase-service-account.json"
	}

	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found: %s", credFile)
	}

	opt := option.WithCredentialsFile(credFile)
	return firebase.NewApp(ctx, cfg, opt)
}

// setEmulatorVars ensures emulator env vars are set
func setEmulatorVars() {
	defaults := map[string]string{
		"FIRESTORE_EMULATOR_HOST":        "localhost:8080",
		"FIREBASE_AUTH_EMULATOR_HOST":    "localhost:9099",
		"FIREBASE_STORAGE_EMULATOR_HOST": "localhost:9199",
	}

	for key, val := range defaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
