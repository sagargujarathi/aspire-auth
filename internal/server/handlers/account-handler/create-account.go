package account

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	var req request.CreateAccountRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v\nBody: %s", err, string(c.Body()))
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid JSON format in request body",
		})
	}

	// Debug log
	log.Printf("Received request: %+v", req)

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" ||
		req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "All required fields must be provided",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error hashing password",
		})
	}

	var dateOfBirth *time.Time
	if req.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return c.Status(400).JSON(response.APIResponse{
				Success: false,
				Message: "Invalid date format. Use DD-MM-YYYY",
			})
		}
		dateOfBirth = &parsedDate
	}

	var gender *models.GenderType
	if req.Gender != "" {
		g := models.GenderType(req.Gender)
		gender = &g
	}

	// Handle file upload
	var avatarFilename *string
	if req.Avatar != nil {
		filename, err := helpers.SaveFile(req.Avatar, "images/avatars")
		if err != nil {
			log.Printf("Error saving avatar: %v", err)
			return c.Status(500).JSON(response.APIResponse{
				Success: false,
				Message: "Error saving avatar",
			})
		}
		avatarFilename = &filename
	}

	account := models.Account{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dateOfBirth,
		Gender:         gender,
		Avatar:         avatarFilename,
		RoleType:       models.RoleUser,
	}

	if err := h.DB.Create(&account).Error; err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error creating account",
		})
	}

	// Generate OTP
	otp := fmt.Sprintf("%06d", rand.Intn(999999))

	// Store OTP in Redis
	err = h.Redis.Set(c.Context(), fmt.Sprintf("otp:%s", account.ID.String()), otp, 15*time.Minute).Err()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error in verification system. Please try again.",
		})
	}

	// Send verification email
	if err := helpers.SendVerificationEmail(account.Email, otp, h.Config); err != nil {
		log.Printf("Email error: %v", err)
		return c.Status(201).JSON(response.CreateAccountResponse{
			APIResponse: response.APIResponse{
				Success: true,
				Message: "Account created successfully, but email verification failed. Please contact support.",
			},
			AccountID: account.ID.String(),
		})
	}

	return c.Status(201).JSON(response.CreateAccountResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Account created successfully. Please check your email for verification.",
		},
		AccountID: account.ID.String(),
	})
}
