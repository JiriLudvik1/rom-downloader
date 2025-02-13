package main

import (
	"context"
	"fmt"
	"log"
	"rom-downloader/config"
	"rom-downloader/storage/gcs"
	"rom-downloader/storage/local"
	"rom-downloader/subscribing"
)

func main() {
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx := context.Background()
	gcsClient := gcs.NewClient(ctx, configuration)
	fsClient := local.NewFsClient(configuration)

	messages := make(chan subscribing.RomUploadedMessage, 10)
	go subscribing.StartSubscriber(
		ctx,
		configuration,
		messages)

	for message := range messages {
		localFilePath, err := gcsClient.DownloadFile(message.File)
		if err != nil {
			log.Printf("Error downloading file %s: %v", message.File, err)
			return
		}
		fmt.Printf("Downloaded file %s\n", message.File)
		err = fsClient.ProcessLocalFile(localFilePath)
		if err != nil {
			log.Printf("Error processing file %s: %v", message.File, err)
		}
	}
	return
}
