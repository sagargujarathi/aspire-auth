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

// CreateAccount godoc
// @Summary Create a new user account
// @Description Creates a new user account and sends verification email
// @Tags accounts
// @Accept json
// @Produce json
// @Param account body request.CreateAccountRequest true "Account information"
// @Success 201 {object} response.CreateAccountResponse
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /account [post]
func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
				Message: "Invalid date format. Use YYYY-MM-DD",
			})
		}
		dateOfBirth = &parsedDate
	}

	var gender *models.GenderType
	if req.Gender != "" {
		g := models.GenderType(req.Gender)
		gender = &g
	}

	var avatarFilename *string
	var avatarPath string

	// Handle avatar base64
	if c.Get("Avatar") != "" { // Check if Avatar key exists in the request
		if req.Avatar != "" { // Check if Avatar value is not empty
			filename, err := helpers.SaveBase64File(req.Avatar, "images/avatars", ".png")
			if err != nil {
				tx.Rollback()
				log.Printf("Error saving avatar: %v", err)
				return c.Status(500).JSON(response.APIResponse{
					Success: false,
					Message: "Error saving avatar",
				})
			}
			avatarFilename = &filename
			avatarPath = fmt.Sprintf("images/avatars/%s", filename)

			defer func() {
				if tx.Error != nil {
					if err := helpers.DeleteFile(avatarPath); err != nil {
						log.Printf("Failed to delete avatar file after error: %v", err)
					}
				}
			}()
		}
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

	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
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
		tx.Rollback()
		log.Printf("Redis error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error in verification system. Please try again.",
		})
	}

	// Send verification email
	if err := helpers.SendVerificationEmail(account.Email, otp, h.Config); err != nil {
		tx.Rollback()
		log.Printf("Email error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error sending verification email. Please try again.",
		})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("Transaction commit error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error completing account creation. Please try again.",
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
