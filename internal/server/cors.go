package server

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func setupCORS(app *fiber.App) {
	// Get allowed origins from environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:4200,http://localhost:3000" // Default allowed origins
	}

	log.Printf("Setting up CORS with allowed origins: %s", allowedOrigins)

	// Create CORS configuration
	corsConfig := cors.Config{
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		MaxAge:       86400, // 24 hours in seconds
	}

	// If we're in development mode, we'll need to handle credentials and origins carefully
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "development" {
		log.Println("Running in development mode, configuring CORS for development")
		// If we need credentials, we can't use wildcard origins
		corsConfig.AllowCredentials = true
		corsConfig.AllowOrigins = allowedOrigins
	} else {
		// In production, we're more strict
		corsConfig.AllowCredentials = true
		corsConfig.AllowOrigins = allowedOrigins
	}

	app.Use(cors.New(corsConfig))
}
