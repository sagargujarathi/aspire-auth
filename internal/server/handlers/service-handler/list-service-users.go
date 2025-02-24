package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *ServiceHandler) ListServiceUsers(c *fiber.Ctx) error {
	var req request.ServiceUsersListRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	var serviceUsers []models.ServicesUser
	var total int64

	query := h.DB.Model(&models.ServicesUser{}).Where("service_id = ?", req.ServiceID)

	if err := query.Count(&total).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching service users",
		})
	}

	if err := query.Preload("User").
		Offset((req.Page - 1) * req.Limit).
		Limit(req.Limit).
		Find(&serviceUsers).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching service users",
		})
	}

	userResponses := make([]response.ServiceUserResponse, len(serviceUsers))
	for i, su := range serviceUsers {
		userResponses[i] = response.ServiceUserResponse{
			ID:         su.ID.String(),
			Username:   su.User.Username,
			Email:      su.User.Email,
			IsVerified: su.IsVerified,
			JoinedAt:   su.CreatedAt,
		}
	}

	return c.Status(200).JSON(response.ServiceUsersListResponse{
		APIResponse: response.APIResponse{
			Success: true,
			Message: "Service users fetched successfully",
		},
		Users: userResponses,
		Total: total,
	})
}
