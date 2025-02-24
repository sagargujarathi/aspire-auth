package helpers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveFile(file *multipart.FileHeader, directory string) (string, error) {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.ReplaceAll(strings.ToLower(filepath.Base(file.Filename)), " ", "_"), ext)
	filepath := filepath.Join(directory, filename)

	// Save the file
	if err := os.MkdirAll(filepath, 0755); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}
