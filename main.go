package main

import (
	"context"
	"fmt"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
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

	existingFileNames, err := getExistingFileNames(destionationDir)
	if err != nil {
		log.Fatalf("Error getting existing file names: %v", err)
	}

	for _, file := range fileList.Items {
		fmt.Printf("Found file %s \n", file.OriginalFilename)
		if contains(existingFileNames, file.OriginalFilename) {
			fmt.Printf("File %s already exists\n", file.OriginalFilename)
			continue
		}

		if downloadFile(srv, file) != nil {
			log.Printf("Error downloading file %s: %v\n", file.OriginalFilename, err)
			continue
		}
		existingFileNames = append(existingFileNames, file.OriginalFilename)
		log.Printf("Downloaded file %s \n", file.OriginalFilename)
	}
}

func getExistingFileNames(folderPath string) ([]string, error) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Folder does not exist: %s", folderPath)
	}

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading directory: %v", err)
	}
	var existingFileNames []string
	for _, file := range files {
		if !file.IsDir() {
			existingFileNames = append(existingFileNames, file.Name())
		}
	}
	return existingFileNames, nil
}

func downloadFile(srv *drive.Service, file *drive.File) error {
	resp, err := srv.Files.Get(file.Id).Download()
	if err != nil {
		return fmt.Errorf("Error downloading file: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destionationDir + "\\" + file.OriginalFilename)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Error copying file contents: %v", err)
	}
	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
