package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) DeleteService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AccountAuthorizationToken)

	// Check ownership
	var service models.Service
	if err := h.DB.Where("id = ? AND owner_id = ?", serviceID, authToken.UserID).First(&service).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Service not found or not authorized")
	}

	// Delete related records
	if err := h.DB.Where("service_id = ?", serviceID).Delete(&models.ServicesUser{}).Error; err != nil {
		return utils.HandleError(c, err)
	}

	// Delete service
	if err := h.DB.Delete(&service).Error; err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Service deleted successfully", nil)
}
