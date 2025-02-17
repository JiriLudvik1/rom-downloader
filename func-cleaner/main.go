package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", CleanupHandler)

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func CleanupHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	log.Printf("Starting cleanup...")

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Firestore client: %v", err), http.StatusInternalServerError)
		return
	}
	defer firestoreClient.Close()

	// Initialize Cloud Storage client
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Cloud Storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer storageClient.Close()

	// Get the Firestore collection name from environment variable
	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		http.Error(w, "FIRESTORE_COLLECTION environment variable is not set", http.StatusInternalServerError)
		return
	}

	// Fetch all documents without DeletedAt
	collection := firestoreClient.Collection(collectionName)
	query := collection.Where("deletedAt", "==", nil) // Retrieve documents where DeletedAt is nil
	iter := query.Documents(ctx)

	totalDeleted := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading Firestore documents: %v", err), http.StatusInternalServerError)
			return
		}

		// Map data to CompleteDownload structure
		var data struct {
			MessageId  string `firestore:"messageId"`
			FileName   string `firestore:"fileName"`
			BucketName string `firestore:"bucketName"`
			DeletedAt  *time.Time
		}
		if err := doc.DataTo(&data); err != nil {
			http.Error(w, fmt.Sprintf("Error mapping document data: %v", err), http.StatusInternalServerError)
			return
		}

		// Delete the file from Cloud Storage
		err = deleteFileFromStorage(ctx, storageClient, data.BucketName, data.FileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete file %s/%s: %v", data.BucketName, data.FileName, err), http.StatusInternalServerError)
			return
		}

		// Mark Firestore document as deleted (add DeletedAt timestamp)
		now := time.Now().UTC()
		_, err = doc.Ref.Update(ctx, []firestore.Update{
			{Path: "DeletedAt", Value: now},
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update Firestore document %s: %v", doc.Ref.ID, err), http.StatusInternalServerError)
			return
		}

		totalDeleted++
	}

	// Return success response
	log.Printf("Successfully deleted %d documents and corresponding files.", totalDeleted)
	fmt.Fprintf(w, "Successfully deleted %d documents and corresponding files.\n", totalDeleted)
}

// Helper function to delete a file from Cloud Storage
func deleteFileFromStorage(ctx context.Context, client *storage.Client, bucketName, fileName string) error {
	bucket := client.Bucket(bucketName)
	object := bucket.Object(fileName)

	// Deletes the file
	err := object.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete file from Cloud Storage: %w", err)
	}

	return nil
}
