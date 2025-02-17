package auth

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req request.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	// Remove "Bearer " prefix if present
	tokenString := req.RefreshToken
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Verify refresh token
	authToken := &models.AuthorizationToken{}

	if err := helpers.ParseAccessToken(tokenString, authToken); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	if authToken.TokenType != "ACCOUNT" {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
	}

	// Check if refresh token exists in database
	var refreshTokenModel models.RefreshToken
	if err := h.DB.Where("refresh_token = ? AND expires_at > ?", tokenString, time.Now()).First(&refreshTokenModel).Error; err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
		})
	}

	// Generate new access token
	userID := authToken.UserID
	roleType := authToken.RoleType

	newAccessTokenClaims := &jwt.MapClaims{
		"user_id":    userID,
		"token_type": "ACCOUNT",
		"role_type":  roleType,
		"expires_at": time.Now().Add(time.Minute * 15).Unix(),
	}

	newAccessToken, err := helpers.GenerateAccessToken(newAccessTokenClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating new access token",
		})
	}

	// Generate new refresh token
	newRefreshTokenClaims := &jwt.MapClaims{
		"user_id":    userID,
		"token_type": "ACCOUNT",
		"role_type":  roleType,
		"expires_at": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	newRefreshToken, err := helpers.GenerateRefreshToken(newRefreshTokenClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating new refresh token",
		})
	}

	refreshTokenModel = models.RefreshToken{
		UserID:       uuid.MustParse(userID),
		TokenType:    "SERVICE",
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
	}

	if err := h.DB.Save(&refreshTokenModel).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error updating refresh token",
		})
	}

	return c.Status(200).JSON(response.LoginResponse{
		AccessToken:  "Bearer " + newAccessToken,
		RefreshToken: "Bearer " + newRefreshToken,
	})
}
