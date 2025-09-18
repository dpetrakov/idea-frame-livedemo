package domain

import (
	"errors"
	"time"
)

// User представляет пользователя системы
type User struct {
	ID           string    `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	DisplayName  string    `json:"displayName" db:"display_name"`
	PasswordHash string    `json:"-" db:"password_hash"` // не сериализуется в JSON
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// UserRegister - DTO для регистрации пользователя
type UserRegister struct {
	Login           string `json:"login"`
	DisplayName     string `json:"displayName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// UserLogin - DTO для входа пользователя
type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse - DTO ответа после аутентификации
type AuthResponse struct {
	User      *User     `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserBrief - краткая информация о пользователе
type UserBrief struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
}

// ToUserBrief конвертирует User в UserBrief
func (u *User) ToUserBrief() *UserBrief {
	return &UserBrief{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: u.DisplayName,
	}
}

// Доменные ошибки
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
