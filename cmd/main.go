package main

import (
	"aspire-auth/internal/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	DB_URL := os.Getenv("DB_CONNECTION_URL")
	PORT_NUMBER := os.Getenv("PORT_ADDRESS")

	if DB_URL == "" || PORT_NUMBER == "" {
		log.Fatal("DB_CONNECTION_URL environment variable is not set")
	}

	server := server.NewAPIServer(DB_URL, PORT_NUMBER)

	server.InitHandlers()

	server.Run()
}
