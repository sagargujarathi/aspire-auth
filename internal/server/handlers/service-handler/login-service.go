package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *ServiceHandler) LoginService(c *fiber.Ctx) error {
	var req request.LoginServiceRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.Email == "" || req.Password == "" || req.ServiceID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Email, password and service ID are required")
	}

	// Check if service exists
	var service models.Service
	if err := h.DB.Where("id = ?", req.ServiceID).First(&service).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Service not found")
	}

	var account models.Account
	var serviceUser models.ServicesUser

	// First find the account
	if err := h.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Account not found")
	}

	// Then check if user is associated with the service
	if err := h.DB.Where("user_id = ? AND service_id = ?", account.ID, req.ServiceID).
		First(&serviceUser).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "User not associated with this service")
	}

	// Check password before verification status to avoid timing attacks
	if err := bcrypt.CompareHashAndPassword([]byte(account.HashedPassword), []byte(req.Password)); err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Check both account and service-specific verification
	if !account.IsVerified || !serviceUser.IsVerified {
		return utils.SendError(c, fiber.StatusUnauthorized, "Account or service access not verified")
	}

	claims := &jwt.MapClaims{
		"user_id":    serviceUser.ID.String(),
		"role_type":  account.RoleType,
		"service_id": service.ID.String(),
		"expires_at": time.Now().Add(time.Minute * 15).Unix(),
	}

	accessToken, err := h.Container.JWT.GenerateServiceAccessToken(claims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating access token",
		})
	}

	refreshClaims := &jwt.MapClaims{
		"user_id":    serviceUser.ID.String(),
		"role_type":  account.RoleType,
		"service_id": service.ID.String(),
		"expires_at": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	refreshToken, err := h.Container.JWT.GenerateServiceAccessToken(refreshClaims)
	if err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error generating refresh token",
		})
	}

	refreshTokenModel := models.RefreshToken{
		UserID:       account.ID,
		RefreshToken: refreshToken,
		ServiceID:    &service.ID,
	}
	if err := h.DB.Create(&refreshTokenModel).Error; err != nil {
		log.Printf("Error saving refresh token: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error saving refresh token",
		})
	}

	return c.Status(200).JSON(response.LoginServiceResponse{
		Success:      true,
		Message:      "Service logged in successfully",
		RefreshToken: "Bearer " + refreshToken,
		AccessToken:  "Bearer " + accessToken,
	})

}
