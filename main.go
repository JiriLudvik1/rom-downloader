package main

import (
	"context"
	"fmt"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
	"path/filepath"
	"rom-downloader/config"
	"rom-downloader/subscribing"
	"strings"
	"sync"
)

func main() {
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx := context.Background()
	subscribing.StartSubscriber(
		&ctx,
		configuration)
	return
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(configuration.CredentialsFileName))
	if err != nil {
		log.Fatal(err)
	}

	fileList, err := srv.Files.List().Q("'" + configuration.GoogleDriveFolderId + "' in parents").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	existingFileNames, err := getExistingFileNames(configuration.DestinationFolderRoot)
	if err != nil {
		log.Fatalf("Error getting existing file names: %v", err)
	}

	wg := &sync.WaitGroup{}

	for _, file := range fileList.Items {
		fmt.Printf("Found file %s \n", file.Title)

		// Check if the file name already exists locally
		if fileExistsInList(existingFileNames, file.Title) {
			fmt.Printf("File %s already exists\n", file.Title)
			continue
		}
		wg.Add(1)
		go executeDownload(file, wg, configuration, srv)
	}
	wg.Wait()
	log.Printf("Done downloading files\n")
}

func executeDownload(file *drive.File, wg *sync.WaitGroup, config *config.LoaderConfig, srv *drive.Service) {
	extension, err := getFileExtension(file)
	if err != nil {
		fmt.Printf("Error getting file extension: %v", err)
		wg.Done()
		return
	}

	// Find the appropriate destination folder
	typeFolder, exists := config.RomTypeDestinations[extension]
	if !exists {
		log.Printf("Skipping file %s: No folder configured for %s files\n", file.Title, extension)
		wg.Done()
		return
	}
	destFolder := filepath.Join(config.DestinationFolderRoot, typeFolder)

	// Ensure the destination folder exists
	if err := os.MkdirAll(destFolder, os.ModePerm); err != nil {
		log.Fatalf("Error creating destination folder %s: %v\n", destFolder, err)
	}

	// Download the file to the specific folder
	if downloadFileToFolder(srv, file, destFolder) != nil {
		log.Printf("Error downloading file %s: %v\n", file.Title, err)
		wg.Done()
		return
	}

	log.Printf("Downloaded and sorted file %s into folder %s\n", file.Title, destFolder)
	wg.Done()
}

func getFileExtension(file *drive.File) (string, error) {
	if underscoreIndex := strings.LastIndex(file.Title, "_"); underscoreIndex != -1 {
		// Find the portion between the last "_" and the "."
		dotIndex := strings.LastIndex(file.Title, ".")
		if dotIndex > underscoreIndex {
			// Extract and convert to uppercase
			return strings.ToUpper(file.Title[underscoreIndex+1 : dotIndex]), nil
		} else {
			// Fallback: extension cannot be determined correctly
			log.Printf("Skipping file %s: Invalid format (no valid '.' or '_')\n", file.Title)
			return "", fmt.Errorf("Skipping file %s: Invalid format (no valid '.' or '_')\n", file.Title)
		}
	} else {
		// No underscore found, log or skip
		log.Printf("Skipping file %s: No valid underscore ('_') found\n", file.Title)
		return "", fmt.Errorf("Skipping file %s: Invalid format (no valid '.' or '_')\n", file.Title)
	}
}

func getExistingFileNames(folderPath string) ([]string, error) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Folder does not exist: %s", folderPath)
	}

	var existingFileNames []string

	var traverse func(string) error
	traverse = func(path string) error {
		files, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("Error reading directory %s: %v", path, err)
		}

		for _, file := range files {
			fullPath := filepath.Join(path, file.Name())
			if file.IsDir() {
				if err := traverse(fullPath); err != nil {
					return err
				}
			} else {
				existingFileNames = append(existingFileNames, fullPath)
			}
		}
		return nil
	}

	if err := traverse(folderPath); err != nil {
		return nil, err
	}

	return existingFileNames, nil
}

func downloadFileToFolder(srv *drive.Service, file *drive.File, destFolder string) error {
	resp, err := srv.Files.Get(file.Id).Download()
	if err != nil {
		return fmt.Errorf("Error downloading file: %v", err)
	}
	defer resp.Body.Close()

	filePath := filepath.Join(destFolder, file.Title)

	out, err := os.Create(filePath)
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

func fileExistsInList(existingFileNames []string, targetFileName string) bool {
	for _, fullPath := range existingFileNames {
		// Extract the file name from the full path
		if filepath.Base(fullPath) == targetFileName {
			return true
		}
	}
	return false
}
