package account

import (
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(account *models.Account) error
	FindByID(id string) (*models.Account, error)
	FindByEmail(email string) (*models.Account, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	Verify(id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *repository) FindByID(id string) (*models.Account, error) {
	var account models.Account
	if err := r.db.First(&account, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *repository) FindByEmail(email string) (*models.Account, error) {
	var account models.Account
	if err := r.db.First(&account, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *repository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.Account{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(id string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete related records
	if err := tx.Where("user_id = ?", id).Delete(&models.ServicesUser{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("user_id = ?", id).Delete(&models.RefreshToken{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.Account{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *repository) Verify(id string) error {
	return r.db.Model(&models.Account{}).Where("id = ?", id).Update("is_verified", true).Error
}
