package server

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type APIServer struct {
	databaseURL string
	portAddress string
	app         *fiber.App
	database    *gorm.DB
}

func NewAPIServer(databaseURL string, portAddress string) *APIServer {
	app := fiber.New()
	database, err := gorm.Open(postgres.Open(databaseURL))

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return &APIServer{app: app, database: database, databaseURL: databaseURL, portAddress: portAddress}
}

func (s *APIServer) InitHandlers() {
	fmt.Println("Initializing handlers...")
}

func (s *APIServer) Run() {

	log.Fatal(s.app.Listen(s.portAddress))
}
