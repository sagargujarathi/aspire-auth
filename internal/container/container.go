package container

import (
	"aspire-auth/internal/app"
	"aspire-auth/internal/config"
	"aspire-auth/internal/domain/account"
	"aspire-auth/internal/domain/service"
	"aspire-auth/internal/validation"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	Config         *config.Config
	DB             *gorm.DB
	Redis          *redis.Client
	AccountUseCase *app.AccountUseCase
	AuthUseCase    *app.AuthUseCase
	ServiceUseCase *app.ServiceUseCase
}

func NewContainer(cfg *config.Config, db *gorm.DB, redis *redis.Client) *Container {
	// Initialize validator
	validator := validation.NewValidator()

	// Initialize repositories
	accountRepo := account.NewRepository(db)
	serviceRepo := service.NewRepository(db)

	// Initialize use cases
	accountUseCase := app.NewAccountUseCase(accountRepo, validator, redis)
	authUseCase := app.NewAuthUseCase(accountRepo, cfg)
	serviceUseCase := app.NewServiceUseCase(serviceRepo)

	return &Container{
		Config:         cfg,
		DB:             db,
		Redis:          redis,
		AccountUseCase: &accountUseCase,
		AuthUseCase:    authUseCase,
		ServiceUseCase: serviceUseCase,
	}
}
