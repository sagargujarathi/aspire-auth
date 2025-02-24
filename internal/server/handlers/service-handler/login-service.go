package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
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

	var userRoleType models.RoleType = "USER"
	if service.OwnerID == account.ID {
		userRoleType = "ADMIN"
	}

	accountTokenClaims := &models.ServiceRefreshToken{
		UserID:    account.ID,
		RoleType:  userRoleType,
		ServiceID: service.ID,
		ExpiresAt: time.Now().Add(h.Config.JWT.Account.AccessExpiry),
	}

	refreshTokenModel := models.ServiceRefreshToken{
		UserID:    account.ID,
		ServiceID: service.ID,
		RoleType:  userRoleType,
		ExpiresAt: time.Now().Add(h.Config.JWT.Account.RefreshExpiry),
	}

	accessToken, err := h.Container.JWT.GenerateServiceAccessToken(accountTokenClaims)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating access token")
	}

	refreshToken, err := h.Container.JWT.GenerateServiceRefreshToken(accountTokenClaims)
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating refresh token")
	}

	refreshTokenModel.RefreshToken = refreshToken

	if err := h.DB.Create(&refreshTokenModel).Error; err != nil {
		log.Printf("Error saving refresh token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error saving refresh token")
	}

	return c.Status(200).JSON(response.LoginServiceResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Service logged in successfully",
		},
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	})

}
