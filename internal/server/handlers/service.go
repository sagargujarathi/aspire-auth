package handlers

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/response"
	"aspire-auth/internal/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ServiceHandler struct {
	*BaseHandler
}

func NewServiceHandler(base *BaseHandler) *ServiceHandler {
	return &ServiceHandler{BaseHandler: base}
}

func (h *ServiceHandler) handleError(c *fiber.Ctx, err error) error {
	log.Printf("Database error: %v", err)
	return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error")
}

func (h *ServiceHandler) CreateService(c *fiber.Ctx) error {
	var req request.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	// Get owner's ID from auth token
	authToken := c.Locals("auth").(*models.AuthorizationToken)
	ownerID, err := uuid.Parse(authToken.ID)
	if err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	service := models.Service{
		OwnerID:            ownerID,
		ServiceName:        req.ServiceName,
		ServiceDescription: req.ServiceDescription,
		ServiceLogo:        req.ServiceLogo,
	}

	if err := h.DB.Create(&service).Error; err != nil {
		log.Printf("Error creating service: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error creating service",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success":    true,
		"message":    "Service created successfully",
		"service_id": service.ID,
	})
}

func (h *ServiceHandler) SignupToService(c *fiber.Ctx) error {
	var req request.SignupToServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	authToken := c.Locals("auth").(*models.AuthorizationToken)
	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid service ID",
		})
	}

	userID, err := uuid.Parse(authToken.ID)
	if err != nil {
		return c.Status(400).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Check if user is already signed up
	var existingSignup models.ServicesUser
	if err := h.DB.Where("service_id = ? AND user_id = ?", serviceID, userID).First(&existingSignup).Error; err == nil {
		return utils.SendError(c, fiber.StatusConflict, "Already signed up to this service")
	}

	serviceUser := models.ServicesUser{
		ServiceID: serviceID,
		UserID:    userID,
	}

	if err := h.DB.Create(&serviceUser).Error; err != nil {
		log.Printf("Error signing up to service: %v", err)
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error signing up to service",
		})
	}

	return c.Status(200).JSON(response.APIResponse{
		Success: true,
		Message: "Successfully signed up to service",
	})
}

func (h *ServiceHandler) ListMyServices(c *fiber.Ctx) error {
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	var services []models.Service
	var total int64

	if err := h.DB.Model(&models.Service{}).Where("owner_id = ?", authToken.ID).Count(&total).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching services",
		})
	}

	if err := h.DB.Where("owner_id = ?", authToken.ID).Find(&services).Error; err != nil {
		return c.Status(500).JSON(response.APIResponse{
			Success: false,
			Message: "Error fetching services",
		})
	}

	serviceResponses := make([]response.ServiceResponse, len(services))
	for i, service := range services {
		var usersCount int64
		h.DB.Model(&models.ServicesUser{}).Where("service_id = ?", service.ID).Count(&usersCount)

		serviceResponses[i] = response.ServiceResponse{
			ID:          service.ID.String(),
			Name:        service.ServiceName,
			Description: service.ServiceDescription,
			Logo:        service.ServiceLogo,
			UsersCount:  usersCount,
		}
	}

	return c.Status(200).JSON(response.ServiceListResponse{
		Success:  true,
		Message:  "Services fetched successfully",
		Services: serviceResponses,
		Total:    total,
	})
}

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
			UserID:     su.UserID.String(),
			Username:   su.User.Username,
			Email:      su.User.Email,
			IsVerified: su.IsVerified,
			JoinedAt:   su.CreatedAt,
		}
	}

	return c.Status(200).JSON(response.ServiceUsersListResponse{
		Success: true,
		Message: "Service users fetched successfully",
		Users:   userResponses,
		Total:   total,
	})
}

func (h *ServiceHandler) DeleteService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	// Check ownership
	var service models.Service
	if err := h.DB.Where("id = ? AND owner_id = ?", serviceID, authToken.ID).First(&service).Error; err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Service not found or not authorized")
	}

	// Delete related records
	if err := h.DB.Where("service_id = ?", serviceID).Delete(&models.ServicesUser{}).Error; err != nil {
		return h.handleError(c, err)
	}

	// Delete service
	if err := h.DB.Delete(&service).Error; err != nil {
		return h.handleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Service deleted successfully", nil)
}

func (h *ServiceHandler) LeaveService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	// Delete service user record
	result := h.DB.Where("service_id = ? AND user_id = ?", serviceID, authToken.ID).Delete(&models.ServicesUser{})
	if result.Error != nil {
		return h.handleError(c, result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.SendError(c, fiber.StatusNotFound, "You are not a member of this service")
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Successfully left the service", nil)
}
