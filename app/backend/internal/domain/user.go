package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Login       string    `json:"login" db:"login"`
	DisplayName string    `json:"displayName" db:"display_name"`
	PasswordHash string   `json:"-" db:"password_hash"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// UserRegisterRequest - запрос на регистрацию
type UserRegisterRequest struct {
	Login           string `json:"login"`
	DisplayName     string `json:"displayName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// UserLoginRequest - запрос на вход
type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse - ответ после успешной аутентификации
type AuthResponse struct {
	User      *User     `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Доменные ошибки
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUnauthorized      = errors.New("unauthorized")
)