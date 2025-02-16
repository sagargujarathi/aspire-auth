package main

import (
	"aspire-auth/internal/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize server
	server, err := server.NewAPIServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
