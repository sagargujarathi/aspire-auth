package handlers

import (
	"aspire-auth/internal/container"
	"aspire-auth/internal/server/handlers/account-handler"
	"aspire-auth/internal/server/handlers/auth-handler"
	"aspire-auth/internal/server/handlers/service-handler"
)

type Handlers struct {
	Account *account.AccountHandler
	Auth    *auth.AuthHandler
	Service *service.ServiceHandler
}

func InitHandlers(container *container.Container) *Handlers {

	return &Handlers{
		Account: account.NewAccountHandler(container),
		Auth:    auth.NewAuthHandler(container),
		Service: service.NewServiceHandler(container),
	}
}
