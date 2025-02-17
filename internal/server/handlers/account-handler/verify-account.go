package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

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
