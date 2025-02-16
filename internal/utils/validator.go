package utils

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return NewServiceError(400, "Validation failed", validationErrors)
	}
	return nil
}

func ValidateEmail(email string) bool {
	return validate.Var(email, "required,email") == nil
}

func ValidatePassword(password string) bool {
	// At least 8 chars, 1 upper, 1 lower, 1 number, 1 special
	return validate.Var(password, "required,min=8,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789,containsany=!@#$%^&*()") == nil
}
