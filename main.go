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
	"strings"
)

const (
	credentialsFileName = "service-account.json"
	romFolderID         = "1i-SG3ixBSt5dZPn3ihSD9BsWnzgWUfJm"
	destionationDir     = "C:\\testDir"
)

func main() {
	romTypeDestinations := map[string]string{
		"NES":  filepath.Join(destionationDir, "NES"),
		"N64":  filepath.Join(destionationDir, "N64"),
		"SNES": filepath.Join(destionationDir, "SNES"),
		"GB":   filepath.Join(destionationDir, "GameBoy"),
	}

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
		fmt.Printf("Found file %s \n", file.Title)

		// Check if the file name already exists locally
		if fileExistsInList(existingFileNames, file.Title) {
			fmt.Printf("File %s already exists\n", file.Title)
			continue
		}

		var extension string
		if underscoreIndex := strings.LastIndex(file.Title, "_"); underscoreIndex != -1 {
			// Find the portion between the last "_" and the "."
			dotIndex := strings.LastIndex(file.Title, ".")
			if dotIndex > underscoreIndex {
				// Extract and convert to uppercase
				extension = strings.ToUpper(file.Title[underscoreIndex+1 : dotIndex])
			} else {
				// Fallback: extension cannot be determined correctly
				log.Printf("Skipping file %s: Invalid format (no valid '.' or '_')\n", file.Title)
				continue
			}
		} else {
			// No underscore found, log or skip
			log.Printf("Skipping file %s: No valid underscore ('_') found\n", file.Title)
			continue
		}

		// Find the appropriate destination folder
		destFolder, exists := romTypeDestinations[extension]
		if !exists {
			log.Printf("Skipping file %s: No folder configured for %s files\n", file.Title, extension)
			continue
		}

		// Ensure the destination folder exists
		if err := os.MkdirAll(destFolder, os.ModePerm); err != nil {
			log.Fatalf("Error creating destination folder %s: %v\n", destFolder, err)
		}

		// Download the file to the specific folder
		if downloadFileToFolder(srv, file, destFolder) != nil {
			log.Printf("Error downloading file %s: %v\n", file.Title, err)
			continue
		}

		// Add the file to the list of downloaded files
		existingFileNames = append(existingFileNames, filepath.Join(destFolder, file.Title))
		log.Printf("Downloaded and sorted file %s into folder %s\n", file.Title, destFolder)
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
