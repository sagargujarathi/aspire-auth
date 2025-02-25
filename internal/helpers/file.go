package helpers

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
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
	filename := fmt.Sprintf("%d_%s%s",
		time.Now().UnixNano(),
		strings.ReplaceAll(strings.ToLower(filepath.Base(file.Filename)), " ", "_"),
		ext)
	filePath := filepath.Join(directory, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}

func SaveBase64File(base64Data, directory, extension string) (string, error) {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	filePath := filepath.Join(directory, filename)

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}

func DeleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func IsValidImagePath(directory, filename string) bool {
	// Validate directory (only allow specific directories)
	allowedDirs := map[string]bool{
		"avatars": true,
		// Add other allowed directories here
	}

	if !allowedDirs[directory] {
		log.Printf("Invalid directory requested: %s", directory)
		return false
	}

	// Validate filename (prevent directory traversal)
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		log.Printf("Invalid filename (contains path separators): %s", filename)
		return false
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !allowedExts[ext] {
		log.Printf("Invalid file extension: %s", ext)
		return false
	}

	return true
}
