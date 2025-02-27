package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type FirestoreEntry struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Set up the HTTP server.
	http.HandleFunc("/", HandlePubSubPush)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// HandlePubSubPush processes HTTP requests with Pub/Sub messages.
func HandlePubSubPush(w http.ResponseWriter, r *http.Request) {
	log.Println("Received HTTP request for Pub/Sub push")

	// Read the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		log.Printf("Error reading request body: %v", err)
		return
	}
	defer r.Body.Close()

	// Parse the Pub/Sub message from the request body.
	var pubSubMessage PubSubMessage
	if err := json.Unmarshal(body, &pubSubMessage); err != nil {
		http.Error(w, "Invalid Pub/Sub message format", http.StatusBadRequest)
		log.Printf("Error unmarshaling Pub/Sub message: %v", err)
		return
	}

	// Decode the "Data" field which contains the actual message.
	var message string
	if err := json.Unmarshal(pubSubMessage.Data, &message); err != nil {
		http.Error(w, "Invalid message data format", http.StatusBadRequest)
		log.Printf("Error unmarshaling message data: %v", err)
		return
	}

	log.Printf("Received message: %s", message)

	// Process the message using Firestore.
	ctx := context.Background()
	if err := writeToFirestore(ctx, message); err != nil {
		http.Error(w, "Failed to process message", http.StatusInternalServerError)
		log.Printf("Error writing to Firestore: %v", err)
		return
	}

	// Respond to the Pub/Sub system to confirm receipt.
	w.WriteHeader(http.StatusOK)
	log.Println("Message successfully processed")
}

// writeToFirestore writes the message data to Firestore.
func writeToFirestore(ctx context.Context, message string) error {
	// Retrieve Firestore project and collection configuration from environment variables.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return fmt.Errorf("environment variable GOOGLE_CLOUD_PROJECT is not set")
	}
	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		return fmt.Errorf("environment variable FIRESTORE_COLLECTION is not set")
	}

	// Initialize Firestore client.
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %w", err)
	}
	defer client.Close()

	// Create a Firestore entry.
	entry := FirestoreEntry{
		Message:   message,
		Timestamp: time.Now(),
	}

	// Write the entry to Firestore.
	_, _, err = client.Collection(collectionName).Add(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to write to Firestore: %w", err)
	}

	return nil
}
