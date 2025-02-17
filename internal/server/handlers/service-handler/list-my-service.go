package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) ListMyServices(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	var services []models.Service
	var total int64

	if err := h.DB.Model(&models.Service{}).Where("owner_id = ?", authToken.ID).Count(&total).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching services",
		})
	}

	if err := h.DB.Where("owner_id = ?", authToken.ID).Find(&services).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching services",
		})
	}

	serviceResponses := make([]response.ServiceResponse, len(services))
	for i, service := range services {
		var usersCount int64
		h.DB.Model(&models.ServicesUser{}).Where("service_id = ?", service.ID).Count(&usersCount)

		serviceResponses[i] = response.ServiceResponse{
			ID:          service.ID.String(),
			Name:        service.ServiceName,
			Description: service.ServiceDescription,
			Logo:        service.ServiceLogo,
			UsersCount:  usersCount,
		}
	}

	return c.Status(200).JSON(response.ServiceListResponse{
		Success:  true,
		Message:  "Services fetched successfully",
		Services: serviceResponses,
		Total:    total,
	})
}
