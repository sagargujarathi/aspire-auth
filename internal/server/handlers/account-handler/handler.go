package account

import "aspire-auth/internal/container"

type AccountHandler struct {
	*container.Container
}

func NewAccountHandler(base *container.Container) *AccountHandler {
	return &AccountHandler{Container: base}
}
