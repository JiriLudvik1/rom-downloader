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
	"rom-downloader/storage/local"
)

type Client struct {
	storageClient *storage.Client
	context       context.Context
	config        *config.LoaderConfig
}

func NewGcsClient(ctx context.Context, config *config.LoaderConfig) *Client {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(config.CredentialsFileName))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return &Client{
		storageClient: client,
		context:       ctx,
		config:        config,
	}
}

func (g *Client) DownloadFile(fileName string) (string, error) {
	destinationFilePath := filepath.Join(g.config.TempFolder, fileName)
	if local.FileExists(destinationFilePath) {
		log.Printf("File %s already exists, skipping download", destinationFilePath)
		return destinationFilePath, nil
	}

	destinationDir := filepath.Dir(destinationFilePath)
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", destinationDir, err)
	}

	bucket := g.storageClient.Bucket(g.config.BucketName)
	obj := bucket.Object(fileName)

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file %s: %w", destinationFilePath, err)
	}

	defer func() {
		if err := destinationFile.Close(); err != nil {
			log.Printf("Error closing destination file %s: %v", destinationFilePath, err)
		}
	}()

	reader, err := obj.NewReader(g.context)
	if err != nil {
		return "", fmt.Errorf("failed to create reader for file %s: %w", fileName, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader for file %s: %v", fileName, err)
		}
	}()

	copied, err := g.copyWithCancellation(destinationFile, reader)
	if err != nil {
		return "", fmt.Errorf("failed to copy file %s: %w", fileName, err)
	}

	log.Printf("Successfully copied %d bytes for file %s", copied, fileName)

	return destinationFilePath, nil
}

func (g *Client) copyWithCancellation(dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64

	for {
		select {
		case <-g.context.Done():
			return written, g.context.Err()
		default:
		}

		nr, readErr := src.Read(buf)
		if nr > 0 {
			nw, writeErr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}

			if writeErr != nil {
				return written, writeErr
			}

			if nr != nw {
				return written, io.ErrShortWrite
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return written, readErr
		}
	}

	return written, nil
}

func (g *Client) Close() error {
	return g.storageClient.Close()
}
