package middleware

import (
	"aspire-auth/internal/container"
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Middleware struct {
	*container.Container
}

func InitMiddleware(container *container.Container) *Middleware {
	return &Middleware{container}
}

func (h *Middleware) AccountAuthMiddleware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
	}

	token := h.Container.JWT.ExtractToken(authorization)
	authToken := &models.AuthorizationToken{}

	if err := h.Container.JWT.ParseAccountAccessToken(token, authToken); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	c.Locals("auth", authToken)
	return c.Next()
}

func (h *Middleware) ServiceAuthMiddlerware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
	}

	token := h.Container.JWT.ExtractToken(authorization)
	authToken := &models.AuthorizationToken{}

	if err := h.Container.JWT.ParseServiceAccessToken(token, authToken); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	c.Locals("auth", authToken)
	return c.Next()
}

func ContentType(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.Next()
}
