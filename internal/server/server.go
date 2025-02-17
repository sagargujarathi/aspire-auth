package server

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/container"
	"aspire-auth/internal/middleware"
	"aspire-auth/internal/server/handlers"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type APIServer struct {
	container *container.Container
	app       *fiber.App
	handlers  *handlers.Handlers
}

func NewAPIServer() *APIServer {
	cfg := config.Load()

	db := initDatabase(cfg)

	redis := initRedis(cfg)
	app := initFiber(cfg)
	container := container.NewContainer(cfg, db, redis, app)

	h := handlers.InitHandlers(container)

	return &APIServer{
		container: container,
		app:       app,
		handlers:  h,
	}
}

func initDatabase(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	return db
}

func initRedis(config *config.Config) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        config.Redis.Address,
		Password:    config.Redis.Password,
		DB:          config.Redis.DB,
		DialTimeout: 5 * time.Second,
		MaxRetries:  3,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Continuing without Redis - OTP functionality will not work")
	} else {
		log.Println("Successfully connected to Redis")
	}

	return redisClient
}

func initFiber(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		JSONDecoder:  json.Unmarshal,
		JSONEncoder:  json.Marshal,
	})

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Add JSON content type middleware
	app.Use(middleware.ContentType)

	return app
}

func (s *APIServer) InitHandlers() {
	// Public routes
	s.app.Post("/account", s.handlers.Account.CreateAccount)
	s.app.Post("/verify", s.handlers.Account.VerifyAccount)
	s.app.Post("/login", s.handlers.Auth.Login)
	s.app.Post("/refresh-token", s.handlers.Auth.RefreshToken)
	s.app.Post("/resend-otp", s.handlers.Account.ResendOTP)

	protected := s.app.Group("", middleware.AuthMiddleware)

	// Protected routes
	protected.Put("/account", s.handlers.Account.UpdateAccount)
	protected.Delete("/account", s.handlers.Account.DeleteAccount)

	// Service routes

	protected.Post("/service", s.handlers.Service.CreateService)
	protected.Put("/service/:id", s.handlers.Service.UpdateService)
	protected.Get("/service/my", s.handlers.Service.ListMyServices)
	protected.Post("/service/signup", s.handlers.Service.SignupToService)
	protected.Post("/service/users", s.handlers.Service.ListServiceUsers)
	protected.Delete("/service/:id", s.handlers.Service.DeleteService)
	protected.Delete("/service/:id/leave", s.handlers.Service.LeaveService)
}

func (s *APIServer) Run() error {
	s.InitHandlers()
	log.Printf("Server is running on port %s", s.container.Config.Server.Port)
	return s.app.Listen(s.container.Config.Server.Port)
}
