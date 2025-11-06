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

// InitFirebase initializes Firebase app and its clients
func InitFirebase() error {
	ctx := context.Background()

	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credPath == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("failed to create Firebase App: %w", err)
	}

	if AuthClient, err = app.Auth(ctx); err != nil {
		return fmt.Errorf("failed to init Auth: %w", err)
	}

	if FirestoreClient, err = app.Firestore(ctx); err != nil {
		return fmt.Errorf("failed to init Firestore: %w", err)
	}

	log.Println("✅ Firebase initialized successfully")
	return nil
}
