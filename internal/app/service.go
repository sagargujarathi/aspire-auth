package app

import (
	"aspire-auth/internal/domain/service"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/utils"
	"context"

	"github.com/google/uuid"
)

type ServiceUseCase struct {
	repo service.Repository
}

func NewServiceUseCase(repo service.Repository) *ServiceUseCase {
	return &ServiceUseCase{repo: repo}
}

func (uc *ServiceUseCase) Create(ctx context.Context, req *request.CreateServiceRequest, ownerID string) (*models.Service, error) {
	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, utils.NewServiceError(400, "Invalid owner ID", err)
	}

	service := &models.Service{
		OwnerID:            ownerUUID,
		ServiceName:        req.ServiceName,
		ServiceDescription: req.ServiceDescription,
		ServiceLogo:        req.ServiceLogo,
	}

	if err := uc.repo.Create(service); err != nil {
		return nil, utils.NewServiceError(500, "Failed to create service", err)
	}

	return service, nil
}

func (uc *ServiceUseCase) AddUser(ctx context.Context, serviceID, userID string) error {
	svcID, err := uuid.Parse(serviceID)
	if err != nil {
		return utils.NewServiceError(400, "Invalid service ID", err)
	}

	usrID, err := uuid.Parse(userID)
	if err != nil {
		return utils.NewServiceError(400, "Invalid user ID", err)
	}

	return uc.repo.AddUser(svcID, usrID)
}

func (uc *ServiceUseCase) RemoveUser(ctx context.Context, serviceID, userID string) error {
	return uc.repo.RemoveUser(serviceID, userID)
}

func (uc *ServiceUseCase) ListUsers(ctx context.Context, serviceID string, page, limit int) ([]models.ServicesUser, int64, error) {
	return uc.repo.ListUsers(serviceID, page, limit)
}

func (uc *ServiceUseCase) Delete(ctx context.Context, serviceID, ownerID string) error {
	service, err := uc.repo.FindByID(serviceID)
	if err != nil {
		return utils.ErrNotFound
	}

	if service.OwnerID.String() != ownerID {
		return utils.ErrUnauthorized
	}

	return uc.repo.Delete(serviceID)
}
