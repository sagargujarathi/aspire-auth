package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) GetServiceUserDetails(context *fiber.Ctx) error {
	authToken := context.Locals("auth").(*models.ServiceAuthorizationToken)

	serviceUser := models.ServicesUser{}
	if err := h.DB.Model(&models.ServicesUser{}).Where("service_id = ? AND id = ?", authToken.ServiceID, authToken.UserID).First(&serviceUser).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching service user",
		})
	}

	account := models.Account{}
	if err := h.DB.Model(&models.Account{}).Where("id = ?", serviceUser.UserID).First(&account).Error; err != nil {
		return context.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching account",
		})
	}

	return context.Status(200).JSON(response.GetAccountDetailsResponse{
		Success: true,
		Message: "Service user fetched successfully",
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
