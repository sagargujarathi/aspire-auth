package auth

import "aspire-auth/internal/container"

type AuthHandler struct {
	*container.Container
}

func NewAuthHandler(base *container.Container) *AuthHandler {
	return &AuthHandler{
		Container: base,
	}
}
