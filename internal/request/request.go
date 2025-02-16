package request

import (
	"time"
)

type CreateAccountRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

type UpdateAccountRequest struct {
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Avatar      string    `json:"avatar"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerifyAccountRequest struct {
	AccountID string `json:"account_id"`
	OTP       string `json:"otp"`
}

type CreateServiceRequest struct {
	ServiceName        string  `json:"service_name" validate:"required"`
	ServiceDescription *string `json:"service_description,omitempty"`
	ServiceLogo        *string `json:"service_logo,omitempty"`
}

type UpdateServiceRequest struct {
	ServiceName        string  `json:"service_name"`
	ServiceDescription *string `json:"service_description,omitempty"`
	ServiceLogo        *string `json:"service_logo,omitempty"`
}

type SignupToServiceRequest struct {
	ServiceID string `json:"service_id" validate:"required"`
}

type ServiceUsersListRequest struct {
	ServiceID string `json:"service_id" validate:"required"`
	Page      int    `json:"page" default:"1"`
	Limit     int    `json:"limit" default:"10"`
}

type ResendOTPRequest struct {
	AccountID string `json:"account_id" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
