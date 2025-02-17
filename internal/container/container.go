package container

import (
	"aspire-auth/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	app    *fiber.App
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client, app *fiber.App) *Container {

	return &Container{
		Config: cfg,
		DB:     db,
		Redis:  redis,
		app:    app,
	}
}
