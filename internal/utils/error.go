package utils

import (
	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	if serviceErr, ok := err.(*ServiceError); ok {
		return SendError(c, serviceErr.Code, serviceErr.Message)
	}
	return SendError(c, fiber.StatusInternalServerError, "Internal server error")
}
