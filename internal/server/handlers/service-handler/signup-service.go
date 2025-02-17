package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *ServiceHandler) SignupToService(c *fiber.Ctx) error {
	var req request.SignupToServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	if req.ServiceID == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Service ID is required")
	}

	authToken := c.Locals("auth").(*models.AuthorizationToken)
	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid service ID")
	}

	if authToken.TokenType != "ACCOUNT" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	// Check if service exists
	var service models.Service
	if err := h.DB.Where("id = ?", serviceID).First(&service).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Service not found")
	}

	userID, err := uuid.Parse(authToken.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	// Check if user is already signed up
	var existingSignup models.ServicesUser
	if err := h.DB.Where("service_id = ? AND user_id = ?", serviceID, userID).First(&existingSignup).Error; err == nil {
		return utils.SendError(c, fiber.StatusConflict, "Already signed up to this service")
	}

	SERVICE_USER_ID := uuid.New()

	serviceUser := models.ServicesUser{
		ID:        SERVICE_USER_ID,
		ServiceID: serviceID,
		UserID:    userID,
	}

	if err := h.DB.Create(&serviceUser).Error; err != nil {
		log.Printf("Error signing up to service: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error signing up to service",
		})
	}

	return c.Status(200).JSON(response.SignUpServiceResponse{
		Success:       true,
		Message:       "Successfully signed up to service",
		ServiceUserID: SERVICE_USER_ID,
	})
}
