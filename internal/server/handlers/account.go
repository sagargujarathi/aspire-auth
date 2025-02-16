package handlers

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AccountHandler struct {
	*BaseHandler
}

func NewAccountHandler(base *BaseHandler) *AccountHandler {
	return &AccountHandler{BaseHandler: base}
}

func (h *AccountHandler) handleError(c *fiber.Ctx, err error) error {
	log.Printf("Database error: %v", err)
	return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error")
}

func (h *AccountHandler) UpdateAccount(c *fiber.Ctx) error {
	var req request.UpdateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request: %v", err)
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	authToken := c.Locals("auth").(*models.AuthorizationToken)

	account := models.Account{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if req.Gender != "" {
		gender := models.GenderType(req.Gender)
		account.Gender = &gender
	}

	if req.Avatar != "" {
		account.Avatar = &req.Avatar
	}

	if !req.DateOfBirth.IsZero() {
		account.DateOfBirth = &req.DateOfBirth
	}

	if err := h.DB.Model(&models.Account{}).Where("id = ?", authToken.ID).Updates(&account).Error; err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Account updated successfully",
	})
}

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
		parsedDate, err := time.Parse("02-01-2006", req.DateOfBirth)
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

	var avatar *string
	if req.Avatar != "" {
		avatar = &req.Avatar
	}

	account := models.Account{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dateOfBirth,
		Gender:         gender,
		Avatar:         avatar,
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
	if err := helpers.SendVerificationEmail(account.Email, otp); err != nil {
		log.Printf("Email error: %v", err)
		return c.Status(201).JSON(response.APIResponse{
			Success: true,
			Message: "Account created successfully, but email verification failed. Please contact support.",
		})
	}

	return c.Status(201).JSON(response.CreateAccountResponse{
		Success:   true,
		Message:   "Account created successfully. Please check your email for verification.",
		AccountID: account.ID.String(),
	})
}

func (h *AccountHandler) VerifyAccount(c *fiber.Ctx) error {
	var req request.VerifyAccountRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing verify request: %v", err)
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	log.Printf("Verifying account with ID: %s and OTP: %s", req.AccountID, req.OTP)

	// First check if account exists
	var account models.Account
	if err := h.DB.Where("id = ?", req.AccountID).First(&account).Error; err != nil {
		log.Printf("Account not found: %v", err)
		return c.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Account not found",
		})
	}

	// Check OTP in Redis
	redisKey := fmt.Sprintf("otp:%s", req.AccountID)
	storedOTP, err := h.Redis.Get(c.Context(), redisKey).Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP",
		})
	}

	if storedOTP != req.OTP {
		log.Printf("OTP mismatch. Got: %s, Expected: %s", req.OTP, storedOTP)
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid OTP",
		})
	}

	// Update verification status
	if err := h.DB.Model(&account).Update("is_verified", true).Error; err != nil {
		log.Printf("Failed to update verification status: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Failed to verify account",
		})
	}

	// Clean up Redis
	h.Redis.Del(c.Context(), redisKey)

	return c.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Account verified successfully",
	})
}

func (h *AccountHandler) ResendOTP(c *fiber.Ctx) error {
	var req request.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	var account models.Account
	if err := h.DB.Where("id = ?", req.AccountID).First(&account).Error; err != nil {
		return c.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Account not found",
		})
	}

	if account.IsVerified {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Account is already verified",
		})
	}

	// Generate new OTP
	otp := fmt.Sprintf("%06d", rand.Intn(999999))

	// Store in Redis
	err := h.Redis.Set(c.Context(), fmt.Sprintf("otp:%s", account.ID.String()), otp, 15*time.Minute).Err()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating new OTP",
		})
	}

	// Send new verification email
	if err := helpers.SendVerificationEmail(account.Email, otp); err != nil {
		log.Printf("Email error: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error sending verification email",
		})
	}

	return c.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "New OTP sent successfully",
	})
}

func (h *AccountHandler) DeleteAccount(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	// Delete related records first
	if err := h.DB.Where("user_id = ?", authToken.ID).Delete(&models.ServicesUser{}).Error; err != nil {
		return h.handleError(c, err)
	}

	if err := h.DB.Where("user_id = ?", authToken.ID).Delete(&models.RefreshToken{}).Error; err != nil {
		return h.handleError(c, err)
	}

	// Delete the account
	if err := h.DB.Delete(&models.Account{}, "id = ?", authToken.ID).Error; err != nil {
		return h.handleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Account deleted successfully", nil)
}
