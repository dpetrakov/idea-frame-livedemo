package domain

import (
	"errors"
	"fmt"
)

// Стандартные ошибки домена
var (
	ErrNotFound       = errors.New("resource not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ValidationError ошибка валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ErrInvalidInput создаёт ошибку валидации
func ErrInvalidInput(msg string) error {
	return ValidationError{Message: msg}
}

// ErrInvalidField создаёт ошибку валидации для конкретного поля
func ErrInvalidField(field, msg string) error {
	return ValidationError{Field: field, Message: msg}
}