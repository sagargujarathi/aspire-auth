package auth

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

	claims := &jwt.MapClaims{
		"user_id":    account.ID.String(),
		"role_type":  account.RoleType,
		"token_type": "ACCOUNT",
		"expires_at": time.Now().Add(time.Minute * 15).Unix(),
	}

	accessToken, err := h.Container.JWT.GenerateAccountAccessToken(claims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating access token",
		})
	}

	refreshClaims := &jwt.MapClaims{
		"user_id":    account.ID.String(),
		"role_type":  account.RoleType,
		"token_type": "ACCOUNT",
		"expires_at": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	refreshToken, err := h.Container.JWT.GenerateAccountRefreshToken(refreshClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating refresh token",
		})
	}

	refreshTokenModel := models.RefreshToken{
		UserID:       account.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
	}

	if err := h.DB.Create(&refreshTokenModel).Error; err != nil {
		log.Printf("Error saving refresh token: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error saving refresh token",
		})
	}

	return c.Status(200).JSON(response.LoginResponse{
		AccessToken:  "Bearer " + accessToken,
		RefreshToken: "Bearer " + refreshToken,
	})
}
