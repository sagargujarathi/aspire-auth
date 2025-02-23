package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *AccountHandler) DeleteAccount(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AccountAuthorizationToken)

	// Delete the account
	if err := h.DB.Delete(&models.Account{}, "id = ?", authToken.UserID).Error; err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Account deleted successfully", nil)
}
