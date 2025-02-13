package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"rom-downloader/config"
)

type FsClient struct {
	config *config.LoaderConfig
}

func NewFsClient(config *config.LoaderConfig) *FsClient {
	return &FsClient{config: config}
}

func (c *FsClient) ProcessLocalFile(filePath string) error {
	extensions, err := getFileExtensions(filePath)
	if err != nil {
		return err
	}

	filesToRemove := &[]string{filePath}
	defer func() { removeFiles(*filesToRemove) }()

	// We just want not tagged files let be
	if extensions.CustomExtension == nil {
		return nil
	}

	consoleFolder, err := c.getConsoleFolder(extensions)
	if err != nil {
		return err
	}

	if !fileIsArchive(filePath) {
		err = sortFilesToFolders([]string{filePath}, consoleFolder)
		if err != nil {
			return err
		}
		return nil
	}

	extractedPath := path.Join(c.config.TempFolder, "extracted")
	filePaths, err := ExtractArchive(filePath, extractedPath)
	*filesToRemove = append(*filesToRemove, filePaths...)

	if err != nil {
		return err
	}
	err = sortFilesToFolders(filePaths, consoleFolder)
	if err != nil {
		return err
	}
	return nil
}

func sortFilesToFolders(filePaths []string, consoleFolderPath string) error {
	// Ensure the destination folder exists
	err := os.MkdirAll(consoleFolderPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create console folder: %w", err)
	}

	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)

		destinationPath := filepath.Join(consoleFolderPath, fileName)

		err := os.Rename(filePath, destinationPath)
		if err != nil {
			return fmt.Errorf("failed to move file %s: %w", filePath, err)
		}
	}

	return nil
}

func (c *FsClient) getConsoleFolder(identifier *ConsoleIdentifier) (string, error) {
	consoleFolder, exists := c.config.RomTypeDestinations[*identifier.CustomExtension]
	fullConsoleFolder := filepath.Join(c.config.DestinationFolderRoot, consoleFolder)
	if !exists {
		return "", fmt.Errorf("no destination folder configured for ROM type: %s", *identifier.CustomExtension)

	}
	return fullConsoleFolder, nil
}

func removeFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove file %s: %w", filePath, err)
		}
	}
	return nil
}
