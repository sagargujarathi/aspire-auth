package auth

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"time"

	"github.com/gofiber/fiber/v2"
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
	authToken := &models.AccountAuthorizationToken{}

	if err := h.Container.JWT.ParseAccountRefreshToken(tokenString, authToken); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	// Check if refresh token exists in database
	var refreshTokenModel models.AccountRefreshToken
	if err := h.DB.Where("refresh_token = ? AND expires_at > ?", tokenString, time.Now()).First(&refreshTokenModel).Error; err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
		})
	}

	// Generate new access token
	userID := authToken.UserID
	roleType := authToken.RoleType

	newAccessTokenClaims := &models.AccountRefreshToken{
		UserID:    uuid.MustParse(userID),
		RoleType:  roleType,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	newAccessToken, err := h.Container.JWT.GenerateAccountAccessToken(newAccessTokenClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating new access token",
		})
	}

	// Generate new refresh token
	newRefreshTokenClaims := &models.AccountRefreshToken{
		UserID:    uuid.MustParse(userID),
		RoleType:  roleType,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	newRefreshToken, err := h.Container.JWT.GenerateAccountRefreshToken(newRefreshTokenClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating new refresh token",
		})
	}

	// Update refresh token in database
	refreshTokenModel.RefreshToken = newRefreshToken
	refreshTokenModel.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)

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
