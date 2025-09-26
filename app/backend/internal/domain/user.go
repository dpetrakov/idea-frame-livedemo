package domain

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя в системе
type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Login           string     `json:"login" db:"login"`
	DisplayName     string     `json:"displayName" db:"display_name"`
	PasswordHash    string     `json:"-" db:"password_hash"` // Не сериализуется в JSON
	Email           string     `json:"email" db:"email"`
	IsAdmin         bool       `json:"isAdmin"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty" db:"email_verified_at"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time  `json:"-" db:"updated_at"`
}

// UserRegisterRequest запрос на регистрацию
type UserRegisterRequest struct {
	Login           string `json:"login" validate:"required,min=3,max=32,alphanum"`
	DisplayName     string `json:"displayName" validate:"required,min=1,max=32"`
	Email           string `json:"email" validate:"required,email"`
	EmailCode       string `json:"emailCode" validate:"required,len=6"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

// UserLoginRequest запрос на вход
type UserLoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// EmailCodeLoginRequest запрос на вход по e-mail коду (passwordless)
type EmailCodeLoginRequest struct {
	Email     string `json:"email"`
	EmailCode string `json:"emailCode"`
}

// AuthResponse ответ с токеном авторизации
type AuthResponse struct {
	User      *User     `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserBrief краткая информация о пользователе
type UserBrief struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Login       string    `json:"login" db:"login"`
	DisplayName string    `json:"displayName" db:"display_name"`
}

// Validate проверяет корректность данных пользователя для регистрации
func (r *UserRegisterRequest) Validate() error {
	if len(r.Login) < 3 || len(r.Login) > 32 {
		return ErrInvalidInput("login must be between 3 and 32 characters")
	}

	// Проверка формата логина (только буквы, цифры, _, -)
	for _, ch := range r.Login {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' || ch == '-') {
			return ErrInvalidInput("login can only contain letters, numbers, underscore and dash")
		}
	}

	if len(r.DisplayName) < 1 || len(r.DisplayName) > 32 {
		return ErrInvalidInput("displayName must be between 1 and 32 characters")
	}

	// email basic validation
	if strings.TrimSpace(r.Email) == "" {
		return ErrInvalidField("email", "email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidField("email", "invalid email format")
	}
	if len(r.EmailCode) != 6 {
		return ErrInvalidField("emailCode", "emailCode must be 6 characters")
	}

	if len(r.Password) < 8 || len(r.Password) > 64 {
		return ErrInvalidInput("password must be between 8 and 64 characters")
	}

	if r.Password != r.ConfirmPassword {
		return ErrInvalidInput("passwords do not match")
	}

	return nil
}

// Validate проверяет корректность данных для входа
func (r *UserLoginRequest) Validate() error {
	if r.Login == "" {
		return ErrInvalidInput("login is required")
	}
	if r.Password == "" {
		return ErrInvalidInput("password is required")
	}
	return nil
}

// Validate проверяет корректность данных для входа по e-mail коду
func (r *EmailCodeLoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrInvalidField("email", "email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidField("email", "invalid email format")
	}
	if len(strings.TrimSpace(r.EmailCode)) != 6 {
		return ErrInvalidField("emailCode", "emailCode must be 6 characters")
	}
	return nil
}

// EmailCodeRequest запрос на отправку кода подтверждения e‑mail
type EmailCodeRequest struct {
	Email string `json:"email"`
}
