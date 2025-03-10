package middleware

import (
	"aspire-auth/internal/container"
	"aspire-auth/internal/models"
	"aspire-auth/internal/response"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // Use this instead of dgrijalva/jwt-go
)

type Middleware struct {
	*container.Container
}

func InitMiddleware(container *container.Container) *Middleware {
	return &Middleware{container}
}

func (h *Middleware) AccountAuthMiddleware(c *fiber.Ctx) error {
	fmt.Println("ACCOUNT AUTH MIDDLEWARE CALLED for path:", c.Path())
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized: missing token",
		})
	}

	token := h.Container.JWT.ExtractToken(authorization)
	fmt.Printf("ACCOUNT Middleware - Received token: %s\n", token) // Debug log

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
	fmt.Println("SERVICE AUTH MIDDLEWARE CALLED for path:", c.Path())
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Unauthorized: missing token",
		})
	}

	// Make sure we remove the Bearer prefix if present
	token := h.Container.JWT.ExtractToken(authorization)

	fmt.Printf("SERVICE Middleware - Received token: %s\n", token) // Debug log

	// Extract token parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
			Success: false,
			Message: "Invalid token format",
		})
	}

	// Try to parse without validation to determine token type
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	rawClaims := jwt.MapClaims{}
	_, _, err := parser.ParseUnverified(token, &rawClaims)

	var serviceSecret string

	if err == nil {
		// Log the claims and check token type
		fmt.Printf("Token claims: %v\n", rawClaims)

		if _, exists := rawClaims["service_id"]; !exists {
			log.Printf("Warning: This appears to be an account token, not a service token")
			return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
				Success: false,
				Message: "Invalid token type: expected service token",
			})
		} else {
			if serviceID, ok := rawClaims["service_id"].(string); ok {
				// Get the service's secret from the database
				var service models.Service
				if err := h.Container.DB.Where("id = ?", serviceID).First(&service).Error; err != nil {
					fmt.Printf("Error finding service: %v\n", err)
					return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
						Success: false,
						Message: "Invalid service ID in token",
					})
				} else {
					// Decrypt the service secret
					decryptedSecret, err := h.Container.JWT.DecryptServiceSecretKey(service.SecretKey)
					if err != nil {
						fmt.Printf("Error decrypting service secret: %v\n", err)
						return c.Status(fiber.StatusInternalServerError).JSON(response.APIResponse{
							Success: false,
							Message: "Error processing service authentication",
						})
					} else {
						serviceSecret = decryptedSecret
						fmt.Printf("Successfully retrieved service secret for service ID: %s\n", serviceID)
					}
				}
			}
		}
	}

	authToken := &models.ServiceAuthorizationToken{}

	// If we have the service secret, validate with it
	if serviceSecret != "" {
		if err := h.Container.JWT.ParseServiceAccessTokenWithSecret(token, authToken, serviceSecret); err == nil {
			// Successfully validated with service secret
			fmt.Printf("Token validated with service-specific secret\n")

			// Validate token expiration
			if err := authToken.Valid(); err != nil {
				if strings.Contains(err.Error(), "token is expired") {
					return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
						Success: false,
						Message: "Token has expired",
					})
				}
				return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
					Success: false,
					Message: "Token validation failed: " + err.Error(),
				})
			}

			c.Locals("auth", authToken)
			return c.Next()
		} else {
			fmt.Printf("Token validation with service-specific secret failed: %v\n", err)

			// Provide specific error messages
			if strings.Contains(err.Error(), "token is expired") {
				return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
					Success: false,
					Message: "Token has expired",
				})
			}

			return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
				Success: false,
				Message: "Invalid token signature",
			})
		}
	}

	// Fall back to global secret only if we couldn't get the service secret
	return c.Status(fiber.StatusUnauthorized).JSON(response.APIResponse{
		Success: false,
		Message: "Unable to validate service token",
	})
}

func ContentType(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json")
	return c.Next()
}
