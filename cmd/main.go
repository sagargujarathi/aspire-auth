package main

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/server"
	"aspire-auth/internal/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// @title Aspire Auth API
// @version 1.0
// @description Authentication and authorization service for third-party applications
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@aspiredev.in

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:4000
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".

func main() {
	// Try loading from different .env locations
	envPaths := []string{
		".env",
		"../.env",
		filepath.Join(os.Getenv("HOME"), "aspire-auth", ".env"),
	}

	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded environment from %s", path)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Warning: Could not load .env file, using environment variables")
	}

	// Override PORT_ADDRESS if PORT environment variable is set
	if port := os.Getenv("PORT"); port != "" && os.Getenv("PORT_ADDRESS") == "" {
		os.Setenv("PORT_ADDRESS", ":"+port)
		log.Printf("Using PORT environment variable: %s", port)
	}

	// Verify environment
	utils.VerifyEnvironmentSetup()

	// Print important configurations for debugging
	log.Printf("Starting server with PORT_ADDRESS: %s", os.Getenv("PORT_ADDRESS"))
	log.Printf("Database URL: %s", maskSensitiveInfo(os.Getenv("DB_CONNECTION_URL")))
	log.Printf("Redis Address: %s", os.Getenv("REDIS_ADDRESS"))

	// Print JWT secrets for debugging (masked)
	log.Printf("Account Access Secret: %s", maskSensitiveInfo(os.Getenv("ACCOUNT_ACCESS_TOKEN_SECRET_KEY")))
	log.Printf("Service Access Secret: %s", maskSensitiveInfo(os.Getenv("SERVICE_ACCESS_TOKEN_SECRET_KEY")))

	// Print a configuration summary
	cfg := config.Load()
	utils.PrintConfigSummary(cfg)

	server := server.NewAPIServer()
	log.Fatalf("Server failed: %v", server.Run())
}

// Mask sensitive information for logging
func maskSensitiveInfo(input string) string {
	if len(input) < 15 {
		return "********"
	}
	return input[:10] + "..." + input[len(input)-5:]
}
