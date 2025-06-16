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

func InitFirebase() error {
	ctx := context.Background()

	// Check if the service account file exists
	credFile := "firebase-service-account.json"
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return fmt.Errorf("credentials file not found: %s", credFile)
	}

	opt := option.WithCredentialsFile(credFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("🔥 Firebase init failed: %v", err)
		return fmt.Errorf("firebase init: %v", err)
	}

	AuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Printf("🔥 Auth client init failed: %v", err)
		return fmt.Errorf("auth init: %v", err)
	}

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Printf("🔥 Firestore init failed: %v", err)
		return fmt.Errorf("firestore init: %v", err)
	}

	log.Println("✅ Firebase successfully initialized")
	return nil
}


