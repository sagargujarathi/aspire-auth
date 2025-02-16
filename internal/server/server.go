package server

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/middleware"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // Update to v5
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type APIServer struct {
	databaseURL string
	portAddress string
	app         *fiber.App
	database    *gorm.DB
	redis       *redis.Client
}

func NewAPIServer(databaseURL string, portAddress string) *APIServer {
	app := fiber.New()
	database, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return &APIServer{
		app:         app,
		database:    database,
		databaseURL: databaseURL,
		portAddress: portAddress,
		redis:       redisClient,
	}
}

func (s *APIServer) ParseBody(context *fiber.Ctx, request any) error {
	if err := context.BodyParser(request); err != nil {
		return context.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return nil
}

func (s *APIServer) GenerateHashedPassword(context *fiber.Ctx, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return string(hashedPassword), nil
}

func (s *APIServer) generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(999999))
}

func (s *APIServer) sendVerificationEmail(email, otp string) error {
	return helpers.SendVerificationEmail(email, otp)
}

func (s *APIServer) CreateAccount(context *fiber.Ctx) error {
	var request request.CreateAccountRequest

	s.ParseBody(context, &request)

	hashedPassword, _ := s.GenerateHashedPassword(context, request.Password)

	account := models.Account{
		Username:       request.Username,
		Email:          request.Email,
		HashedPassword: hashedPassword,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		Gender:         request.Gender,
		DateOfBirth:    request.DateOfBirth,
		Avatar:         request.Avatar,
	}

	if err := s.database.Create(&account).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	// Generate and send OTP
	otp := s.generateOTP()
	err := s.sendVerificationEmail(account.Email, otp)
	if err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Failed to send verification email",
		})
	}

	// Store OTP in Redis with 15 minutes expiration
	err = s.redis.Set(context.Context(), fmt.Sprintf("otp:%s", account.ID.String()), otp, 15*time.Minute).Err()
	if err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Failed to store OTP",
		})
	}

	return context.Status(201).JSON(response.APIResponse{
		Success: true,
		Message: "Account created successfully. Please check your email for verification.",
	})
}

func (s *APIServer) VerifyAccount(context *fiber.Ctx) error {
	var request request.VerifyAccountRequest
	if err := s.ParseBody(context, &request); err != nil {
		return err
	}

	var account models.Account
	if err := s.database.Where("id = ?", request.AccountID).First(&account).Error; err != nil {
		return context.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Account not found",
		})
	}

	// Get OTP from Redis
	storedOTP, err := s.redis.Get(context.Context(), fmt.Sprintf("otp:%s", request.AccountID)).Result()
	if err != nil {
		return context.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP",
		})
	}

	if storedOTP != request.OTP {
		return context.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid OTP",
		})
	}

	// Update account verification status
	if err := s.database.Model(&account).Update("is_verified", true).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Failed to verify account",
		})
	}

	// Delete OTP from Redis
	s.redis.Del(context.Context(), fmt.Sprintf("otp:%s", request.AccountID))

	return context.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Account verified successfully",
	})
}

func (s *APIServer) Login(context *fiber.Ctx) error {
	var request request.LoginRequest
	s.ParseBody(context, &request)

	var account models.Account
	if err := s.database.Where("email = ?", request.Email).First(&account).Error; err != nil {
		return context.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Account not found",
		})
	}

	if !account.IsVerified {
		return context.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Account not verified",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.HashedPassword), []byte(request.Password)); err != nil {
		return context.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid password",
		})
	}

	claims := &jwt.MapClaims{
		"id":        account.ID.String(),
		"role_type": account.RoleType,
		"exp":       time.Now().Add(time.Minute * 15).Unix(),
	}

	ACCESS_TOKEN, err := helpers.GenerateAccessToken(claims)
	if err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating access token",
		})
	}

	refreshClaims := &jwt.MapClaims{
		"id":        account.ID.String(),
		"role_type": account.RoleType,
		"exp":       time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	REFRESH_TOKEN, err := helpers.GenerateRefreshToken(refreshClaims)
	if err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating refresh token",
		})
	}

	refreshToken := models.RefreshToken{
		UserID:       account.ID,
		RefreshToken: REFRESH_TOKEN,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
	}

	if err := s.database.Create(&refreshToken).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error saving refresh token",
		})
	}

	return context.Status(200).JSON(response.LoginResponse{
		AccessToken:  "Bearer " + ACCESS_TOKEN,
		RefreshToken: "Bearer " + REFRESH_TOKEN,
	})
}

func (s *APIServer) UpdateAccount(context *fiber.Ctx) error {
	var request request.UpdateAccountRequest
	s.ParseBody(context, &request)

	// Get auth data from context
	authToken := context.Locals("auth").(*models.AuthorizationToken)

	account := models.Account{
		Username:    request.Username,
		Email:       request.Email,
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Gender:      request.Gender,
		Avatar:      request.Avatar,
		DateOfBirth: request.DateOfBirth,
	}

	if err := s.database.Model(&models.Account{}).Where("id = ?", authToken.ID).Updates(&account).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return context.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Account updated successfully",
	})
}

func (s *APIServer) InitHandlers() {
	s.app.Post("/account", s.CreateAccount)
	s.app.Post("/login", s.Login)
	s.app.Post("/verify", s.VerifyAccount)

	// Protected routes
	s.app.Put("/account", middleware.AuthMiddleware(), s.UpdateAccount)
}

func (s *APIServer) Run() {
	log.Println("Server is running on port", s.portAddress)
	log.Fatal(s.app.Listen(s.portAddress))
}
