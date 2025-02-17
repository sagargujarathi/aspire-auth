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
)

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
	if err := helpers.SendVerificationEmail(account.Email, otp, h.Config); err != nil {
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
