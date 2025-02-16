package validation

import (
	"aspire-auth/internal/request"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Validator interface {
	ValidateCreateAccount(req *request.CreateAccountRequest) error
	ValidateUpdateAccount(req *request.UpdateAccountRequest) error
	ValidateCreateService(req *request.CreateServiceRequest) error
	ValidateUpdateService(req *request.UpdateServiceRequest) error
}

func (v *customValidator) ValidateCreateService(req *request.CreateServiceRequest) error {
	return validate.Struct(req)
}

func (v *customValidator) ValidateUpdateService(req *request.UpdateServiceRequest) error {
	return validate.Struct(req)
}

type customValidator struct{}

func NewValidator() Validator {
	return &customValidator{}
}

func (v *customValidator) ValidateCreateAccount(req *request.CreateAccountRequest) error {
	return validate.Struct(req)
}

func (v *customValidator) ValidateUpdateAccount(req *request.UpdateAccountRequest) error {
	return validate.Struct(req)
}
