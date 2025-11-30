package main

import (
	"context"
	"log"
	"time"

	"github.com/cecvl/art-print-backend/internal/firebase"
	"github.com/cecvl/art-print-backend/internal/processing"
)

func main() {
	ctx := context.Background()
	if err := firebase.InitFirebase(); err != nil {
		log.Fatalf("firebase init failed: %v", err)
	}
	defer firebase.FirestoreClient.Close()

	// run worker loop; return on fatal
	go func() {
		if err := processing.StartWorker(ctx); err != nil {
			log.Fatalf("worker failed: %v", err)
		}
	}()

	// keep main alive
	for {
		time.Sleep(10 * time.Second)
	}
}
