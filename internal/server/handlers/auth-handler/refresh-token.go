package auth

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.RefreshTokenSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid refresh token",
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
	userID := claims["id"].(string)
	roleType := models.RoleType(claims["role_type"].(string))

	newAccessTokenClaims := &jwt.MapClaims{
		"id":        userID,
		"role_type": roleType,
		"exp":       time.Now().Add(time.Minute * 15).Unix(),
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
		"id":        userID,
		"role_type": roleType,
		"exp":       time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	newRefreshToken, err := helpers.GenerateRefreshToken(newRefreshTokenClaims)
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
