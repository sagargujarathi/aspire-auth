package handlers

import (
	"aspire-auth/internal/app"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type ServiceHandler struct {
	useCase *app.ServiceUseCase
}

func NewServiceHandler(useCase *app.ServiceUseCase) *ServiceHandler {
	return &ServiceHandler{useCase: useCase}
}

func (h *ServiceHandler) Create(c *fiber.Ctx) error {
	var req request.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	authToken := c.Locals("auth").(*models.AuthorizationToken)
	service, err := h.useCase.Create(c.Context(), &req, authToken.ID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "Service created successfully", service)
}

func (h *ServiceHandler) SignupToService(c *fiber.Ctx) error {
	var req request.SignupToServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request format")
	}

	authToken := c.Locals("auth").(*models.AuthorizationToken)
	err := h.useCase.AddUser(c.Context(), req.ServiceID, authToken.ID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Successfully signed up to service", nil)
}

func (h *ServiceHandler) LeaveService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	err := h.useCase.RemoveUser(c.Context(), serviceID, authToken.ID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Successfully left service", nil)
}

func (h *ServiceHandler) Delete(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	authToken := c.Locals("auth").(*models.AuthorizationToken)

	err := h.useCase.Delete(c.Context(), serviceID, authToken.ID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Service deleted successfully", nil)
}
