package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *ServiceHandler) RefreshServiceToken(c *fiber.Ctx) error {
	var req request.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Extract token
	tokenString := req.RefreshToken
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Verify refresh token
	authToken := &models.ServiceAuthorizationToken{}
	if err := h.Container.JWT.ParseServiceRefreshToken(tokenString, authToken); err != nil {
		log.Printf("Invalid service refresh token: %v", err)
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid token")
	}

	// Check if refresh token exists in database
	var refreshTokenModel models.ServiceRefreshToken
	if err := h.DB.Where("refresh_token = ? AND expires_at > ?",
		tokenString, time.Now()).First(&refreshTokenModel).Error; err != nil {
		log.Printf("Refresh token not found in database or expired: %v", err)
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid or expired refresh token")
	}

	// Generate new tokens
	userID := uuid.MustParse(authToken.UserID)
	serviceID := uuid.MustParse(authToken.ServiceID)
	roleType := authToken.RoleType

	newTokenModel := &models.ServiceRefreshToken{
		UserID:    userID,
		ServiceID: serviceID,
		RoleType:  roleType,
		ExpiresAt: time.Now().Add(h.Config.JWT.Service.RefreshExpiry),
	}

	// Get the service to retrieve its secret key
	var service models.Service
	if err := h.DB.Where("id = ?", serviceID).First(&service).Error; err != nil {
		log.Printf("Service not found: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error refreshing tokens")
	}

	// Decrypt the service secret key
	serviceSecret, err := h.Container.JWT.DecryptServiceSecretKey(service.SecretKey)
	if err != nil {
		log.Printf("Error decrypting service secret: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error refreshing tokens")
	}

	// Generate new tokens with the service-specific secret
	newAccessToken, err := h.Container.JWT.GenerateServiceAccessTokenWithSecret(newTokenModel, serviceSecret)
	if err != nil {
		log.Printf("Error generating new service access token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating new access token")
	}

	newRefreshToken, err := h.Container.JWT.GenerateServiceRefreshTokenWithSecret(newTokenModel, serviceSecret)
	if err != nil {
		log.Printf("Error generating new service refresh token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating new refresh token")
	}

	// Update refresh token in database
	refreshTokenModel.RefreshToken = newRefreshToken
	refreshTokenModel.ExpiresAt = time.Now().Add(h.Config.JWT.Service.RefreshExpiry)
	if err := h.DB.Save(&refreshTokenModel).Error; err != nil {
		log.Printf("Error updating refresh token in database: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error updating refresh token")
	}

	// Set cookies
	c.Cookie(&fiber.Cookie{
		Name:     "SERVICE_REFRESH_TOKEN",
		Value:    newRefreshToken,
		Path:     "/",
		Expires:  refreshTokenModel.ExpiresAt,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "SERVICE_ACCESS_TOKEN",
		Value:    newAccessToken,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.JWT.Service.AccessExpiry),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "None",
	})

	return c.Status(fiber.StatusOK).JSON(response.LoginServiceResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Service tokens refreshed successfully",
		},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}
