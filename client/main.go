package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"rom-downloader/config"
	"rom-downloader/persistence"
	"rom-downloader/storage/gcs"
	"rom-downloader/storage/local"
	"rom-downloader/subscribing"
	"syscall"
)

func main() {
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	gcsClient := gcs.NewGcsClient(ctx, configuration)
	defer func() {
		if err := gcsClient.Close(); err != nil {
			log.Printf("Error closing GCS client: %v", err)
		}
	}()

	fsClient := local.NewFsClient(configuration)

	firestoreService, err := persistence.NewFirestoreService(ctx, configuration)
	if err != nil {
		log.Fatalf("Error creating firestore service: %v", err)
	}

	messages := make(chan subscribing.RomUploadedMessage, 10)
	go func() {
		subscribing.StartSubscriber(
			ctx,
			configuration,
			messages)
		close(messages)
	}()

	for message := range messages {
		localFilePath, err := gcsClient.DownloadFile(&message)
		if err != nil {
			log.Printf("Error downloading file %s: %v", message.File, err)
			continue
		}
		fmt.Printf("Downloaded file %s\n", message.File)
		err = fsClient.ProcessLocalFile(localFilePath)
		if err != nil {
			log.Printf("Error processing file %s: %v", message.File, err)
		}

		completeDownload := persistence.CompleteDownloadFromMessage(&message)
		err = firestoreService.CreateCompleteDownloadDoc(completeDownload)
		if err != nil {
			log.Printf("Error writing complete download to firestore: %v", err)
		}
	}

	log.Println("Shutting down...")
}
