package container

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	App    *fiber.App
	JWT    *helpers.JWTHelpers
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client, app *fiber.App, jwt *helpers.JWTHelpers) *Container {

	return &Container{
		Config: cfg,
		DB:     db,
		Redis:  redis,
		App:    app,
		JWT:    jwt,
	}
}
