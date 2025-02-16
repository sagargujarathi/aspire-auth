package service

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(service *models.Service) error
	FindByID(id string) (*models.Service, error)
	FindByOwner(ownerID string) ([]models.Service, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	AddUser(serviceID, userID uuid.UUID) error
	RemoveUser(serviceID, userID string) error
	ListUsers(serviceID string, page, limit int) ([]models.ServicesUser, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(service *models.Service) error {
	return r.db.Create(service).Error
}

func (r *repository) FindByID(id string) (*models.Service, error) {
	var service models.Service
	if err := r.db.First(&service, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &service, nil
}

func (r *repository) FindByOwner(ownerID string) ([]models.Service, error) {
	var services []models.Service
	if err := r.db.Where("owner_id = ?", ownerID).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *repository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.Service{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(id string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("service_id = ?", id).Delete(&models.ServicesUser{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.Service{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *repository) AddUser(serviceID, userID uuid.UUID) error {
	serviceUser := models.ServicesUser{
		ServiceID: serviceID,
		UserID:    userID,
	}
	return r.db.Create(&serviceUser).Error
}

func (r *repository) RemoveUser(serviceID, userID string) error {
	return r.db.Where("service_id = ? AND user_id = ?", serviceID, userID).Delete(&models.ServicesUser{}).Error
}

func (r *repository) ListUsers(serviceID string, page, limit int) ([]models.ServicesUser, int64, error) {
	var users []models.ServicesUser
	var total int64

	query := r.db.Model(&models.ServicesUser{}).Where("service_id = ?", serviceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Preload("User").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
