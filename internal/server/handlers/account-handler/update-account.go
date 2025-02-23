package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/utils"
	"log"

	"github.com/gofiber/fiber/v2"
)

func (h *AccountHandler) UpdateAccount(c *fiber.Ctx) error {
	var req request.UpdateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request: %v", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	authToken := c.Locals("auth").(*models.AccountAuthorizationToken)

	account := models.Account{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if req.Gender != "" {
		gender := models.GenderType(req.Gender)
		account.Gender = &gender
	}

	if req.Avatar != "" {
		account.Avatar = &req.Avatar
	}

	if !req.DateOfBirth.IsZero() {
		account.DateOfBirth = &req.DateOfBirth
	}

	if err := h.DB.Model(&models.Account{}).Where("id = ?", authToken.UserID).Updates(&account).Error; err != nil {
		log.Printf("Database error: %v", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update account")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Account updated successfully", nil)
}
