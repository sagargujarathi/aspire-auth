package handlers

import (
	"aspire-auth/internal/app"
	"aspire-auth/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type BaseHandler struct {
	AccountService *app.AccountUseCase
	AuthService    *app.AuthUseCase
	ServiceService *app.ServiceUseCase
}

func (h *BaseHandler) handleError(err error) *fiber.Error {
	if serviceErr, ok := err.(*utils.ServiceError); ok {
		return fiber.NewError(serviceErr.Code, serviceErr.Message)
	}
	return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
}
