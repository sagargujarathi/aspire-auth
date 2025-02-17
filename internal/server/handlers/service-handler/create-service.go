package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *ServiceHandler) CreateService(c *fiber.Ctx) error {
	var req request.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	// Get owner's ID from auth token
	authToken := c.Locals("auth").(*models.AuthorizationToken)
	ownerID, err := uuid.Parse(authToken.ID)
	if err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	service := models.Service{
		OwnerID:            ownerID,
		ServiceName:        req.ServiceName,
		ServiceDescription: req.ServiceDescription,
		ServiceLogo:        req.ServiceLogo,
	}

	if err := h.DB.Create(&service).Error; err != nil {
		log.Printf("Error creating service: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error creating service",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success":    true,
		"message":    "Service created successfully",
		"service_id": service.ID,
	})
}
