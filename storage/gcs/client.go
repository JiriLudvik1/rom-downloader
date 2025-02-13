package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
	"path/filepath"
	"rom-downloader/config"
)

type Client struct {
	storageClient *storage.Client
	context       context.Context
	config        *config.LoaderConfig
}

func NewClient(ctx context.Context, config *config.LoaderConfig) *Client {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(config.CredentialsFileName))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	return &Client{
		storageClient: client,
		context:       ctx,
		config:        config,
	}
}

func (g *Client) DownloadFile(fileName string) (string, error) {
	bucket := g.storageClient.Bucket(g.config.BucketName)
	obj := bucket.Object(fileName)

	destinationFilePath := filepath.Join(g.config.TempFolder, fileName)
	destinationDir := filepath.Dir(destinationFilePath)
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", destinationDir, err)
	}

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file %s: %w", destinationFilePath, err)
	}
	defer destinationFile.Close()

	reader, err := obj.NewReader(g.context)
	if err != nil {
		return "", fmt.Errorf("failed to create reader for file %s: %w", fileName, err)
	}
	defer reader.Close()

	_, err = io.Copy(destinationFile, reader)
	if err != nil {
		return "", fmt.Errorf("failed to copy file %s: %w", fileName, err)
	}
	return destinationFilePath, nil
}
