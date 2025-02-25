package static

import (
	"aspire-auth/internal/container"
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/response"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type StaticHandler struct {
	Container *container.Container
}

func NewStaticHandler(container *container.Container) *StaticHandler {
	return &StaticHandler{
		Container: container,
	}
}

func (h *StaticHandler) ServeImage(c *fiber.Ctx) error {
	directory := c.Params("directory")
	filename := c.Params("filename")

	log.Printf("Requested image path: directory=%s, filename=%s", directory, filename)

	// Validate the path
	if !helpers.IsValidImagePath(directory, filename) {
		log.Printf("Invalid image path requested: %s/%s", directory, filename)
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid image path",
		})
	}

	// Get the absolute path to the image
	imagePath := filepath.Join(".", "images", directory, filename)
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error accessing image",
		})
	}

	log.Printf("Attempting to serve image from: %s", absPath)

	// Check if file exists and get file info
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		log.Printf("Image not found: %s", absPath)
		return c.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Image not found",
		})
	} else if err != nil {
		log.Printf("Error checking file: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error accessing image",
		})
	}

	// Handle If-Modified-Since header
	if modifiedSince := c.Get("If-Modified-Since"); modifiedSince != "" {
		if clientModTime, err := http.ParseTime(modifiedSince); err == nil {
			if fileInfo.ModTime().Unix() <= clientModTime.Unix() {
				return c.SendStatus(304) // Not Modified
			}
		}
	}

	// Handle ETag
	etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(304) // Not Modified
	}

	// Set content type based on file extension
	ext := path.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "image/png" // fallback to png if type cannot be determined
	}

	maxAge := 86400 // 24 hours in seconds

	// Set response headers
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	c.Set("ETag", etag)
	c.Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

	return c.SendFile(absPath)
}
