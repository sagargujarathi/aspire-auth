package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *ServiceHandler) CreateService(c *fiber.Ctx) error {
	var req request.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Get owner's ID from auth token
	authToken := c.Locals("auth").(*models.AccountAuthorizationToken)

	ownerID, err := uuid.Parse(authToken.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	// Validate the secret key - ensure it's at least 16 characters for security
	if len(req.SecretKey) < 16 {
		return utils.SendError(c, fiber.StatusBadRequest, "Service secret key must be at least 16 characters")
	}

	// Encrypt the service secret
	encryptedSecret, err := h.Container.JWT.EncryptServiceSecretKey(req.SecretKey)
	if err != nil {
		log.Printf("Error encrypting service secret key: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error processing service secret key")
	}

	service := models.Service{
		OwnerID:            ownerID,
		ServiceName:        req.ServiceName,
		ServiceDescription: req.ServiceDescription,
		ServiceLogo:        req.ServiceLogo,
		SecretKey:          encryptedSecret,
	}

	if err := h.DB.Create(&service).Error; err != nil {
		log.Printf("Error creating service: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Error creating service")
	}

	// Return success but never return the secret key in the response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":    true,
		"message":    "Service created successfully",
		"service_id": service.ID,
	})
}
