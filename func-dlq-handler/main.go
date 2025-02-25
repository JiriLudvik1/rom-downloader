package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func HandlePubSubMessage(ctx context.Context, m PubSubMessage) error {
	log.Println("Received Pub/Sub message, started processing")
	var message string
	if err := json.Unmarshal(m.Data, &message); err != nil {
		log.Printf("Error unmarshaling Pub/Sub message data: %v", err)
		return err
	}

	log.Printf("Received message: %s", message)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Firestore client: %v", err)
		return err
	}
	defer client.Close()

	entry := FirestoreEntry{
		Message:   message,
		Timestamp: time.Now(),
	}

	// Write to Firestore
	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		log.Printf("FIRESTORE_COLLECTION environment variable is not set")
		return fmt.Errorf("firestore collection name is not set")
	}
	_, _, err = client.Collection(collectionName).Add(ctx, entry)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		return err
	}

	log.Printf("Message successfully written to Firestore collection: %s", collectionName)
	return nil
}
