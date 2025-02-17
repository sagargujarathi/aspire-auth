package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *AccountHandler) DeleteAccount(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	// Delete the account
	if err := h.DB.Delete(&models.Account{}, "id = ?", authToken.ID).Error; err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Account deleted successfully", nil)
}
