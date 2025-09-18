package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/repo"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
)

type AuthService struct {
	userRepo   *repo.UserRepository
	jwtService *security.JWTService
}

func NewAuthService(userRepo *repo.UserRepository, jwtService *security.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *domain.UserRegisterRequest) (*domain.AuthResponse, error) {
	// Валидация входных данных
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Хеширование пароля
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user := &domain.User{
		Login:        req.Login,
		DisplayName:  req.DisplayName,
		PasswordHash: hashedPassword,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Генерация JWT токена
	token, expiresAt, err := s.jwtService.GenerateToken(createdUser.ID, createdUser.Login, createdUser.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		User:      createdUser,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Login выполняет аутентификацию пользователя
func (s *AuthService) Login(ctx context.Context, req *domain.UserLoginRequest) (*domain.AuthResponse, error) {
	// Валидация входных данных
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Поиск пользователя
	user, err := s.userRepo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidPassword // Унифицированная ошибка
		}
		return nil, err
	}

	// Проверка пароля
	if !security.CheckPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidPassword
	}

	// Генерация JWT токена
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

// GetCurrentUser получает текущего пользователя по ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// validateRegisterRequest валидирует запрос на регистрацию
func (s *AuthService) validateRegisterRequest(req *domain.UserRegisterRequest) error {
	// Валидация логина
	if len(req.Login) < 3 || len(req.Login) > 32 {
		return fmt.Errorf("login must be between 3 and 32 characters")
	}

	loginPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !loginPattern.MatchString(req.Login) {
		return fmt.Errorf("login contains invalid characters")
	}

	// Валидация отображаемого имени
	if len(req.DisplayName) < 1 || len(req.DisplayName) > 32 {
		return fmt.Errorf("display name must be between 1 and 32 characters")
	}

	// Валидация пароля
	if len(req.Password) < 8 || len(req.Password) > 64 {
		return fmt.Errorf("password must be between 8 and 64 characters")
	}

	// Проверка подтверждения пароля
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}

// validateLoginRequest валидирует запрос на вход
func (s *AuthService) validateLoginRequest(req *domain.UserLoginRequest) error {
	if req.Login == "" {
		return fmt.Errorf("login is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}