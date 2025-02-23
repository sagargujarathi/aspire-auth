package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *AccountHandler) GetAccountDetails(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AccountAuthorizationToken)

	var account models.Account
	if err := h.DB.Where("id = ?", authToken.UserID).First(&account).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Account not found")
	}

	return c.JSON(response.GetAccountDetailsResponse{
		Success: true,
		Message: "Account details retrieved successfully",
		Account: response.AccountResponse{
			Username:    account.Username,
			Email:       account.Email,
			FirstName:   account.FirstName,
			LastName:    account.LastName,
			DateOfBirth: account.DateOfBirth,
			Gender:      account.Gender,
			RoleType:    account.RoleType,
			Avatar:      account.Avatar,
			CreatedAt:   account.CreatedAt,
			UpdatedAt:   account.UpdatedAt,
		},
	})
}
