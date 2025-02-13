package local

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nwaples/rardecode"
)

type ConsoleIdentifier struct {
	CustomExtension *string // Custom extension between "_" and "."
	FileExtension   string  // Real file extension (after ".")
}

var supportedArchives = map[string]func(string, string, *[]string) error{
	".zip": extractZipWithPaths,
	".tar": extractTarWithPaths,
	".gz":  extractTarGzWithPaths,
	".tgz": extractTarGzWithPaths,
	".rar": extractRarWithPaths,
}

func fileIsArchive(filePath string) bool {
	lowerExt := strings.ToLower(filepath.Ext(filePath))
	_, isArchive := supportedArchives[lowerExt]
	return isArchive
}

func getFileExtensions(filePath string) (*ConsoleIdentifier, error) {
	fileName := filepath.Base(filePath)

	underscoreIndex := strings.LastIndex(fileName, "_")
	if underscoreIndex == -1 {
		log.Printf("Skipping file %s: No valid underscore ('_') found\n", filePath)
		return nil, fmt.Errorf("invalid format: no valid underscore ('_') found in %s", filePath)
	}

	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 || dotIndex <= underscoreIndex {
		log.Printf("Skipping file %s: Invalid format (no valid '.' after '_')\n", filePath)
		return nil, fmt.Errorf("invalid format: no valid '.' after '_' in %s", filePath)
	}

	fileExtension := fileName[dotIndex:]
	if fileExtension == "" {
		log.Printf("Skipping file %s: Empty file extension after '.'\n", filePath)
		return nil, fmt.Errorf("invalid format: empty file extension in %s", filePath)
	}

	customExtension := fileName[underscoreIndex+1 : dotIndex]
	if customExtension == "" {
		return &ConsoleIdentifier{FileExtension: fileExtension, CustomExtension: nil}, nil
	}

	customExtensionPtr := strings.ToUpper(customExtension)
	return &ConsoleIdentifier{
		CustomExtension: &customExtensionPtr,
		FileExtension:   strings.ToLower(fileExtension),
	}, nil
}

func ExtractArchive(archivePath, destinationPath string) ([]string, error) {
	// Ensure the destination path exists
	if err := os.MkdirAll(destinationPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Detect the archive type using file extension
	lowerExt := strings.ToLower(filepath.Ext(archivePath))
	extractFunc, supported := supportedArchives[lowerExt]
	if !supported {
		return nil, fmt.Errorf("unsupported archive format: %s", lowerExt)
	}

	// Extract files and collect their paths
	var extractedFiles []string
	err := extractFunc(archivePath, destinationPath, &extractedFiles) // Call the appropriate extraction function
	return extractedFiles, err
}

func extractZipWithPaths(filePath, destination string, extractedFiles *[]string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		extractedFilePath := filepath.Join(destination, file.Name)

		// Check for directory traversal vulnerability
		if !strings.HasPrefix(extractedFilePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return errors.New("illegal file path in zip (directory traversal attack)")
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, os.ModePerm)
		} else {
			err := extractFileFromZip(file, extractedFilePath)
			if err != nil {
				return err
			}
			*extractedFiles = append(*extractedFiles, extractedFilePath) // Collect file path
		}
	}
	return nil
}

func extractFileFromZip(file *zip.File, destinationPath string) error {
	fileReader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file from zip: %w", err)
	}
	defer fileReader.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, fileReader)
	return err
}

func extractTarWithPaths(filePath, destination string, extractedFiles *[]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	return extractTarContentsWithPaths(file, destination, extractedFiles)
}

func extractTarGzWithPaths(filePath, destination string, extractedFiles *[]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	return extractTarContentsWithPaths(gzipReader, destination, extractedFiles)
}

func extractTarContentsWithPaths(reader io.Reader, destination string, extractedFiles *[]string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		extractedFilePath := filepath.Join(destination, header.Name)

		// Check for directory traversal vulnerability
		if !strings.HasPrefix(extractedFilePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return errors.New("illegal file path in tar (directory traversal attack)")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(extractedFilePath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			destinationFile, err := os.Create(extractedFilePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, tarReader); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
			*extractedFiles = append(*extractedFiles, extractedFilePath) // Collect file path
		default:
			continue
		}
	}

	return nil
}

func extractRarWithPaths(filePath, destination string, extractedFiles *[]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open rar file: %w", err)
	}
	defer file.Close()

	rarReader, err := rardecode.NewReader(file, "")
	if err != nil {
		return fmt.Errorf("failed to create rar reader: %w", err)
	}

	for {
		header, err := rarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read rar entry: %w", err)
		}

		extractedFilePath := filepath.Join(destination, header.Name)

		// Check for directory traversal vulnerability
		if !strings.HasPrefix(extractedFilePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return errors.New("illegal file path in rar (directory traversal attack)")
		}

		if header.IsDir {
			if err := os.MkdirAll(extractedFilePath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			destinationFile, err := os.Create(extractedFilePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, rarReader); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
			*extractedFiles = append(*extractedFiles, extractedFilePath) // Collect file path
		}
	}

	return nil
}
