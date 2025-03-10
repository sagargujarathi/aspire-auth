package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *ServiceHandler) GetServiceUserDetails(context *fiber.Ctx) error {
	authToken := context.Locals("auth").(*models.ServiceAuthorizationToken)

	// Log authentication data for debugging
	log.Printf("Service Auth Token: UserID=%s, ServiceID=%s, RoleType=%s",
		authToken.UserID, authToken.ServiceID, authToken.RoleType)

	// Convert string IDs to UUID
	userID, err := uuid.Parse(authToken.UserID)
	if err != nil {
		log.Printf("Invalid user ID in token: %v", err)
		return utils.SendError(context, fiber.StatusBadRequest, "Invalid user ID in token")
	}

	serviceID, err := uuid.Parse(authToken.ServiceID)
	if err != nil {
		log.Printf("Invalid service ID in token: %v", err)
		return utils.SendError(context, fiber.StatusBadRequest, "Invalid service ID in token")
	}

	// Verify the service user relationship exists
	var serviceUser models.ServicesUser
	if err := h.DB.Where("service_id = ? AND user_id = ?", serviceID, userID).First(&serviceUser).Error; err != nil {
		log.Printf("Service user relationship not found: UserID=%s, ServiceID=%s, Error=%v",
			userID, serviceID, err)
		return utils.SendError(context, fiber.StatusNotFound, "Service user relationship not found")
	}

	// Get the account details
	var account models.Account
	if err := h.DB.Where("id = ?", userID).First(&account).Error; err != nil {
		log.Printf("Account not found: ID=%s, Error=%v", userID, err)
		return utils.SendError(context, fiber.StatusNotFound, "Account not found")
	}

	return context.Status(200).JSON(response.GetAccountDetailsResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Service user fetched successfully",
		},
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
