package auth

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req request.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	var account models.Account
	if err := h.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		return c.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Account not found",
		})
	}

	if !account.IsVerified {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Account not verified",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.HashedPassword), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid password",
		})
	}

	tokenModel := models.AccountRefreshToken{
		UserID:    account.ID,
		RoleType:  account.RoleType,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	accessToken, err := h.Container.JWT.GenerateAccountAccessToken(&tokenModel)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating access token",
		})
	}

	refreshToken, err := h.Container.JWT.GenerateAccountRefreshToken(&tokenModel)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating refresh token",
		})
	}

	tokenModel.RefreshToken = refreshToken
	if err := h.DB.Create(&tokenModel).Error; err != nil {
		log.Printf("Error saving refresh token: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error saving refresh token",
		})
	}

	// Update CORS headers
	origin := c.Get("Origin")
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Credentials", "true")

	// Set cookies with correct settings for persistence
	c.Cookie(&fiber.Cookie{
		Name:     "REFRESH_TOKEN",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.JWT.Account.RefreshExpiry),
		MaxAge:   int(h.Config.JWT.Account.RefreshExpiry.Seconds()),
		Domain:   "localhost",
		Secure:   true,
		HTTPOnly: false,
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "ACCESS_TOKEN",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.JWT.Account.AccessExpiry),
		MaxAge:   int(h.Config.JWT.Account.AccessExpiry.Seconds()),
		Domain:   "localhost",
		Secure:   true,
		HTTPOnly: false,
		SameSite: "None",
	})

	return c.Status(200).JSON(response.LoginResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Login successful",
		},

		ExpiresAt: tokenModel.ExpiresAt.Unix(),
	})
}
