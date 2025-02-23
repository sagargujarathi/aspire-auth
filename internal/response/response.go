package response

import (
	"aspire-auth/internal/models"
	"time"

	"github.com/google/uuid"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateAccountResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	AccountID string `json:"account_id"`
}

type ServiceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	UsersCount  int64   `json:"users_count"`
}

type ServiceListResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Services []ServiceResponse `json:"services"`
	Total    int64             `json:"total"`
}

type ServiceUserResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	JoinedAt   time.Time `json:"joined_at"`
}

type ServiceUsersListResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Users   []ServiceUserResponse `json:"users"`
	Total   int64                 `json:"total"`
}

type LoginServiceResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignUpServiceResponse struct {
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	ServiceUserID uuid.UUID `json:"service_user_id"`
}

type AccountResponse struct {
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	DateOfBirth *time.Time         `json:"date_of_birth,omitempty"`
	Gender      *models.GenderType `json:"gender,omitempty"`
	RoleType    models.RoleType    `json:"role_type"`
	Avatar      *string            `gorm:"type:text" json:"avatar,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type GetAccountDetailsResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Account AccountResponse `json:"account"`
}
