package main

import (
	"context"
	"fmt"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"log"
)

const (
	credentialsFileName = "service-account.json"
	romFolderID         = "1i-SG3ixBSt5dZPn3ihSD9BsWnzgWUfJm"
	destionationDir     = "C:\\testDir"
)

func main() {
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credentialsFileName))
	if err != nil {
		log.Fatal(err)
	}

	fileList, err := srv.Files.List().Q("'" + romFolderID + "' in parents").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	for _, file := range fileList.Items {
		fmt.Printf("Found file %s \n", file.OriginalFilename)
	}
}
