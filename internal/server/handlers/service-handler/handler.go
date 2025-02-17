package service

import "aspire-auth/internal/container"

type ServiceHandler struct {
	*container.Container
}

func NewServiceHandler(base *container.Container) *ServiceHandler {
	return &ServiceHandler{Container: base}
}
