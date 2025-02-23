package middleware

import (
	"aspire-auth/internal/container"
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"
	"fmt"

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
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized: missing token",
		})
	}

	token := h.Container.JWT.ExtractToken(authorization)
	fmt.Printf("Received token: %s\n", token) // Debug log

	authToken := &models.AccountAuthorizationToken{}

	if err := h.Container.JWT.ParseAccountAccessToken(token, authToken); err != nil {
		fmt.Printf("Token parsing error: %v\n", err) // Debug log
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid or expired token: %v", err),
		})
	}

	if err := authToken.Valid(); err != nil {
		fmt.Printf("Token validation error: %v\n", err) // Debug log
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Token validation failed",
		})
	}

	c.Locals("auth", authToken)
	return c.Next()
}

func (h *Middleware) ServiceAuthMiddleware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized: missing token",
		})
	}

	token := h.Container.JWT.ExtractToken(authorization)
	fmt.Printf("Received token: %s\n", token) // Debug log

	authToken := &models.ServiceAuthorizationToken{}

	if err := h.Container.JWT.ParseServiceAccessToken(token, authToken); err != nil {
		fmt.Printf("Token parsing error: %v\n", err) // Debug log
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid or expired token: %v", err),
		})
	}

	if err := authToken.Valid(); err != nil {
		fmt.Printf("Token validation error: %v\n", err) // Debug log
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Token validation failed",
		})
	}

	c.Locals("auth", authToken)
	return c.Next()
}

func ContentType(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.Next()
}
