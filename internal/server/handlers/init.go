package handlers

import (
	"aspire-auth/internal/container"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Account *AccountHandler
	Auth    *AuthHandler
	Service *ServiceHandler
}

func InitHandlers(c *container.Container, app *fiber.App) *Handlers {
	// Create base handler first
	baseHandler := NewBaseHandler(c.DB, c.Redis, app)

	return &Handlers{
		Account: NewAccountHandler(baseHandler),
		Auth:    NewAuthHandler(baseHandler, c.AuthUseCase),
		Service: NewServiceHandler(baseHandler),
	}
}
