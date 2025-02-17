package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) UpdateService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	var req request.UpdateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	var service models.Service
	if err := h.DB.Where("id = ? AND owner_id = ?", serviceID, authToken.ID).First(&service).Error; err != nil {
		return c.Status(404).JSON(response.APIResponse{
			Success: false,
			Message: "Service not found or not authorized",
		})
	}

	updates := map[string]interface{}{}
	if req.ServiceName != "" {
		updates["service_name"] = req.ServiceName
	}
	if req.ServiceDescription != nil {
		updates["service_description"] = req.ServiceDescription
	}
	if req.ServiceLogo != nil {
		updates["service_logo"] = req.ServiceLogo
	}

	if err := h.DB.Model(&service).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error updating service",
		})
	}

	return c.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Service updated successfully",
	})
}
