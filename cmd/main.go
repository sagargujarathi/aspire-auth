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

	server := server.NewAPIServer()

	log.Fatalf("Server failed: %v", server.Run())
}
