package server

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/container"
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/middleware"
	"aspire-auth/internal/server/handlers"
	"aspire-auth/internal/server/handlers/static-handler"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// Remove Swagger imports for now
	// _ "aspire-auth/docs"
	// swagger "github.com/swaggo/fiber-swagger"
)

type APIServer struct {
	container  *container.Container
	app        *fiber.App
	handlers   *handlers.Handlers
	middleware *middleware.Middleware
	static     *static.StaticHandler
}

func NewAPIServer() *APIServer {
	cfg := config.Load()

	db := initDatabase(cfg)

	redis := initRedis(cfg)
	app := initFiber(cfg)
	jwtHelpers := helpers.InitJWTHelpers(cfg)
	container := container.NewContainer(cfg, db, redis, app, jwtHelpers)
	middleWare := middleware.InitMiddleware(container)
	static := static.NewStaticHandler(container)

	h := handlers.InitHandlers(container)

	return &APIServer{
		container:  container,
		app:        app,
		handlers:   h,
		middleware: middleWare,
		static:     static,
	}
}

func initDatabase(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.Database.URL))
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
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Fiber error handler: %v", err)
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Setup CORS with improved configuration
	setupCORS(app)

	// Add JSON content type middleware
	app.Use(middleware.ContentType)

	return app
}

func (s *APIServer) InitHandlers() {

	// Health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0",
		})
	})

	// Public routes
	s.app.Get("/images/:directory/:filename", s.static.ServeImage)
	s.app.Post("/account", s.handlers.Account.CreateAccount)
	s.app.Post("/verify", s.handlers.Account.VerifyAccount)
	s.app.Post("/resend-otp", s.handlers.Account.ResendOTP)
	s.app.Post("/signin", s.handlers.Auth.Login)
	s.app.Post("/refresh-token", s.handlers.Auth.RefreshToken)
	s.app.Post("/service/login", s.handlers.Service.LoginService)
	s.app.Post("/service/signup", s.handlers.Service.SignupToService)
	s.app.Post("/service/refresh-token", s.handlers.Service.RefreshServiceToken)



	// Account protected routes group
	accountGroup := s.app.Group("/account", s.middleware.AccountAuthMiddleware)
	accountGroup.Put("/", s.handlers.Account.UpdateAccount)
	accountGroup.Delete("/", s.handlers.Account.DeleteAccount)
	accountGroup.Get("/", s.handlers.Account.GetAccountDetails)

	// IMPORTANT: Routes that need service auth middleware must come BEFORE routes with account auth middleware
	// Service user routes (protected by service auth)
	s.app.Get("/service/user", s.middleware.ServiceAuthMiddleware, s.handlers.Service.GetServiceUserDetails)
	s.app.Get("/service/user/details", s.middleware.ServiceAuthMiddleware, s.handlers.Service.GetServiceUserDetails)
	serviceUserGroup := s.app.Group("/service-user", s.middleware.ServiceAuthMiddleware)
	serviceUserGroup.Delete("/:id/leave", s.handlers.Service.LeaveService)
	serviceUserGroup.Get("/details", s.handlers.Service.GetServiceUserDetails)

	// Service management routes (protected by account auth)
	// IMPORTANT: These must come AFTER the service auth routes to prevent path conflicts
	serviceManageGroup := s.app.Group("/service", s.middleware.AccountAuthMiddleware)
	serviceManageGroup.Post("/", s.handlers.Service.CreateService)
	serviceManageGroup.Put("/:id", s.handlers.Service.UpdateService)
	serviceManageGroup.Get("/list", s.handlers.Service.ListMyServices)
	serviceManageGroup.Get("/users", s.handlers.Service.ListServiceUsers)
	serviceManageGroup.Delete("/:id", s.handlers.Service.DeleteService)
}

func (s *APIServer) Run() error {
	s.InitHandlers()
	log.Printf("Server is running on port %s", s.container.Config.Server.Port)
	return s.app.Listen(s.container.Config.Server.Port)
}
