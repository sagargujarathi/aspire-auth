package request

import (
	"aspire-auth/internal/models"
	"time"
)

type CreateAccountRequest struct {
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	Password    string             `json:"password"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	DateOfBirth *time.Time         `json:"date_of_birth"`
	Gender      *models.GenderType `json:"gender"`
	Avatar      *string            `json:"avatar"`
}

type UpdateAccountRequest struct {
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	DateOfBirth *time.Time         `json:"date_of_birth"`
	Gender      *models.GenderType `json:"gender"`
	Avatar      *string            `json:"avatar"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyAccountRequest struct {
	AccountID string `json:"account_id"`
	OTP       string `json:"otp"`
}
