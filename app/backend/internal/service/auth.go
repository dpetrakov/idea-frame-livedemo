package service

import (
	"context"
	"errors"
	"time"
	
	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
	"github.com/ideaframe/backend/internal/security"
)

// AuthService сервис аутентификации
type AuthService struct {
	userRepo   *repo.UserRepository
	jwtManager *security.JWTManager
}

// NewAuthService создаёт новый сервис аутентификации
func NewAuthService(db *repo.Database, jwtSecret string, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		userRepo:   repo.NewUserRepository(db),
		jwtManager: security.NewJWTManager(jwtSecret, jwtExpiration),
	}
}

// Register регистрация нового пользователя
func (s *AuthService) Register(ctx context.Context, req *domain.UserRegisterRequest) (*domain.AuthResponse, error) {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Хэширование пароля
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
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
	
	// Сохранение в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, domain.ErrInvalidInput("user with this login already exists")
		}
		return nil, err
	}
	
	// Генерация токена
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	
	// Очистка пароля из ответа
	user.PasswordHash = ""
	
	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Login вход пользователя
func (s *AuthService) Login(ctx context.Context, req *domain.UserLoginRequest) (*domain.AuthResponse, error) {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Поиск пользователя
	user, err := s.userRepo.GetByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	
	// Проверка пароля
	if !security.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}
	
	// Генерация токена
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	
	// Очистка пароля из ответа
	user.PasswordHash = ""
	
	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserByID получение пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Очистка пароля
	user.PasswordHash = ""
	return user, nil
}

// ListUsers получение списка пользователей
func (s *AuthService) ListUsers(ctx context.Context) ([]domain.UserBrief, error) {
	return s.userRepo.List(ctx)
}

// ValidateToken проверка токена
func (s *AuthService) ValidateToken(token string) (*security.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}