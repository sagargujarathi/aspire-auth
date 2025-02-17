package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) LeaveService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	// Delete service user record
	result := h.DB.Where("service_id = ? AND user_id = ?", serviceID, authToken.UserID).Delete(&models.ServicesUser{})
	if result.Error != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Error leaving the service")
	}

	if result.RowsAffected == 0 {
		return utils.SendError(c, fiber.StatusNotFound, "You are not a member of this service")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Successfully left the service", nil)
}
