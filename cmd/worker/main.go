package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	// lightweight HTTP server for Cloud Run health checks
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("worker listening on :%s for health checks", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("health server failed: %v", err)
	}
}
