package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
)

// UserRepo определяет интерфейс для работы с пользователями
type UserRepo interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.UserBrief, error)
}

// AuthService содержит бизнес-логику аутентификации
type AuthService struct {
	userRepo   UserRepo
	jwtService *security.JWTService
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(userRepo UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: security.NewJWTService(jwtSecret),
	}
}

// ValidationError представляет ошибки валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *domain.UserRegister) (*domain.AuthResponse, error) {
	// Валидация
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Проверка что пользователь не существует
	_, err := s.userRepo.GetUserByLogin(ctx, req.Login)
	if err == nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Хеширование пароля
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	now := time.Now()
	user := &domain.User{
		Login:        req.Login,
		DisplayName:  req.DisplayName,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерация JWT
	token, expiresAt, err := s.jwtService.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Login осуществляет вход пользователя
func (s *AuthService) Login(ctx context.Context, req *domain.UserLogin) (*domain.AuthResponse, error) {
	// Валидация
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Поиск пользователя
	user, err := s.userRepo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка пароля
	if !security.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Генерация JWT
	token, expiresAt, err := s.jwtService.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// GetCurrentUser возвращает информацию о текущем пользователе
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetAllUsers возвращает список всех пользователей
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*domain.UserBrief, error) {
	return s.userRepo.GetAllUsers(ctx)
}

// ValidateToken проверяет JWT токен
func (s *AuthService) ValidateToken(tokenString string) (*security.Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// validateRegisterRequest валидирует запрос регистрации
func (s *AuthService) validateRegisterRequest(req *domain.UserRegister) error {
	if len(req.Login) < 3 || len(req.Login) > 32 {
		return ValidationError{Field: "login", Message: "Логин должен быть от 3 до 32 символов"}
	}

	loginRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !loginRegex.MatchString(req.Login) {
		return ValidationError{Field: "login", Message: "Логин может содержать только буквы, цифры, _ и -"}
	}

	if len(req.DisplayName) < 1 || len(req.DisplayName) > 32 {
		return ValidationError{Field: "displayName", Message: "Отображаемое имя должно быть от 1 до 32 символов"}
	}

	if len(req.Password) < 8 || len(req.Password) > 64 {
		return ValidationError{Field: "password", Message: "Пароль должен быть от 8 до 64 символов"}
	}

	if req.Password != req.ConfirmPassword {
		return ValidationError{Field: "confirmPassword", Message: "Пароли не совпадают"}
	}

	return nil
}

// validateLoginRequest валидирует запрос входа
func (s *AuthService) validateLoginRequest(req *domain.UserLogin) error {
	if req.Login == "" {
		return ValidationError{Field: "login", Message: "Логин обязателен"}
	}

	if req.Password == "" {
		return ValidationError{Field: "password", Message: "Пароль обязателен"}
	}

	return nil
}
