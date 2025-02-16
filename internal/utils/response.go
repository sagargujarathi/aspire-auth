package utils

import (
	"aspire-auth/internal/response"
	"log"

	"github.com/gofiber/fiber/v2"
)

func SendSuccess(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func SendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(response.APIResponse{
		Success: false,
		Message: message,
	})
}

func HandleDBError(c *fiber.Ctx, err error, message string) error {
	log.Printf("Database error: %v", err)
	return SendError(c, fiber.StatusInternalServerError, message)
}
