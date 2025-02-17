package middleware

import (
	"aspire-auth/internal/helpers"
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
	}

	token := helpers.ExtractToken(authorization)
	authToken := &models.AuthorizationToken{}

	if err := helpers.ParseAccessToken(token, authToken); err != nil {
		return c.Status(401).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token",
		})
	}

	// Store auth data in context for handlers to use
	c.Locals("auth", authToken)
	return c.Next()
}

func ContentType(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.Next()
}
