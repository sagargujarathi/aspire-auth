package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BaseHandler struct {
	DB    *gorm.DB
	Redis *redis.Client
	App   *fiber.App
}

func NewBaseHandler(db *gorm.DB, redis *redis.Client, app *fiber.App) *BaseHandler {
	return &BaseHandler{
		DB:    db,
		Redis: redis,
		App:   app,
	}
}
