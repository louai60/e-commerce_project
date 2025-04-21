package storage

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage implements a local file storage solution
type LocalStorage struct {
	BasePath string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		BasePath: basePath,
	}, nil
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	URL      string
	PublicID string
}

// Upload saves a file to local storage and returns its URL
func (s *LocalStorage) Upload(data []byte, folder, filename string) (*UploadResult, error) {
	// Create folder if it doesn't exist
	folderPath := filepath.Join(s.BasePath, folder)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	// Generate a unique filename if none provided
	if filename == "" {
		filename = fmt.Sprintf("%d_%s", time.Now().UnixNano(), "image.jpg")
	} else {
		// Sanitize filename
		filename = strings.ReplaceAll(filename, " ", "_")
		filename = strings.ReplaceAll(filename, "/", "_")
		filename = strings.ReplaceAll(filename, "\\", "_")
	}

	// Add timestamp to ensure uniqueness
	ext := filepath.Ext(filename)
	basename := filename[:len(filename)-len(ext)]
	filename = fmt.Sprintf("%s_%d%s", basename, time.Now().UnixNano(), ext)

	// Full path to save the file
	filePath := filepath.Join(folderPath, filename)

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Generate a URL for the file
	// In a real environment, this would be a proper URL
	// For local development, we'll use a relative path
	url := fmt.Sprintf("/uploads/%s/%s", folder, filename)

	return &UploadResult{
		URL:      url,
		PublicID: fmt.Sprintf("%s/%s", folder, filename),
	}, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(publicID string) error {
	filePath := filepath.Join(s.BasePath, publicID)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, consider it deleted
	}

	// Delete the file
	return os.Remove(filePath)
}

// SaveBase64Image saves a base64 encoded image to local storage
func (s *LocalStorage) SaveBase64Image(base64Data, folder, filename string) (*UploadResult, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(base64Data, "data:image") {
		parts := strings.Split(base64Data, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid base64 data URL format")
		}
		base64Data = parts[1]
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Upload the decoded data
	return s.Upload(data, folder, filename)
}

// SaveFromReader saves a file from an io.Reader
func (s *LocalStorage) SaveFromReader(reader io.Reader, folder, filename string) (*UploadResult, error) {
	// Read all data from the reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Upload the data
	return s.Upload(data, folder, filename)
}
