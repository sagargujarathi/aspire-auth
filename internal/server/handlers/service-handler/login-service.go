package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"fmt"
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
		log.Printf("Service not found: ID=%s, Error=%v", req.ServiceID, err)
		return utils.SendError(c, fiber.StatusNotFound, "Service not found")
	}

	var account models.Account
	var serviceUser models.ServicesUser

	// First find the account
	if err := h.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		log.Printf("Account not found: Email=%s, Error=%v", req.Email, err)
		return utils.SendError(c, fiber.StatusNotFound, "Account not found")
	}

	// Then check if user is associated with the service
	if err := h.DB.Where("user_id = ? AND service_id = ?", account.ID, req.ServiceID).
		First(&serviceUser).Error; err != nil {
		log.Printf("User not associated with service: UserID=%s, ServiceID=%s, Error=%v",
			account.ID, req.ServiceID, err)
		return utils.SendError(c, fiber.StatusNotFound, "User not associated with this service")
	}

	// Check password before verification status to avoid timing attacks
	if err := bcrypt.CompareHashAndPassword([]byte(account.HashedPassword), []byte(req.Password)); err != nil {
		log.Printf("Invalid password for user: Email=%s", req.Email)
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Check both account and service-specific verification
	if !account.IsVerified || !serviceUser.IsVerified {
		log.Printf("Account or service not verified: AccountVerified=%v, ServiceUserVerified=%v",
			account.IsVerified, serviceUser.IsVerified)
		return utils.SendError(c, fiber.StatusUnauthorized, "Account or service access not verified")
	}

	var userRoleType models.RoleType = models.RoleUser
	if service.OwnerID == account.ID {
		userRoleType = models.RoleAdmin
	}

	// Retrieve and decrypt the service-specific secret key
	serviceSecret, err := h.Container.JWT.DecryptServiceSecretKey(service.SecretKey)
	if err != nil {
		log.Printf("Error decrypting service secret key: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating service tokens")
	}

	tokenModel := models.ServiceRefreshToken{
		UserID:    account.ID,
		RoleType:  userRoleType,
		ServiceID: service.ID,
		ExpiresAt: time.Now().Add(h.Config.JWT.Service.RefreshExpiry),
	}

	// Generate tokens using the service-specific secret
	accessToken, err := h.Container.JWT.GenerateServiceAccessTokenWithSecret(&tokenModel, serviceSecret)
	if err != nil {
		log.Printf("Error generating service access token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating access token")
	}

	// Debug logging for token generation
	fmt.Printf("Generated service token for user: %s, service: %s, with service-specific secret\n",
		account.ID.String(), service.ID.String())

	// Test token verification before returning it
	testAuthToken := &models.ServiceAuthorizationToken{}
	if err := h.Container.JWT.ParseServiceAccessTokenWithSecret(accessToken, testAuthToken, serviceSecret); err != nil {
		log.Printf("Warning: Generated token fails verification with service secret: %v", err)
	} else {
		log.Printf("Token verification successful with service-specific secret!")
	}

	refreshToken, err := h.Container.JWT.GenerateServiceRefreshTokenWithSecret(&tokenModel, serviceSecret)
	if err != nil {
		log.Printf("Error generating service refresh token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error generating refresh token")
	}

	tokenModel.RefreshToken = refreshToken

	if err := h.DB.Create(&tokenModel).Error; err != nil {
		log.Printf("Error saving refresh token: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error saving refresh token")
	}

	// Set cookies for service authentication
	c.Cookie(&fiber.Cookie{
		Name:     "SERVICE_REFRESH_TOKEN",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.JWT.Service.RefreshExpiry),
		MaxAge:   int(h.Config.JWT.Service.RefreshExpiry.Seconds()),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "SERVICE_ACCESS_TOKEN",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.JWT.Service.AccessExpiry),
		MaxAge:   int(h.Config.JWT.Service.AccessExpiry.Seconds()),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "None",
	})

	return c.Status(200).JSON(response.LoginServiceResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Service logged in successfully",
		},
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	})
}
