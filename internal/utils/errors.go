package utils

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotVerified = errors.New("account not verified")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrDuplicateEntry     = errors.New("duplicate entry")
	ErrNotFound           = errors.New("resource not found")
)

type ServiceError struct {
	Code    int
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewServiceError(code int, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
