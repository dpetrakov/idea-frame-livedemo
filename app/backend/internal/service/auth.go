package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/config"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
	"github.com/ideaframe/backend/internal/security"
	"github.com/ideaframe/backend/internal/telemetry"
)

// AuthService сервис аутентификации
type AuthService struct {
	userRepo   *repo.UserRepository
	emailRepo  *repo.EmailCodeRepository
	jwtManager *security.JWTManager
	cfg        *config.Config
	mailSender MailSender
}

// NewAuthService создаёт новый сервис аутентификации
func NewAuthService(db *repo.Database, cfg *config.Config) *AuthService {
	s := &AuthService{
		userRepo:   repo.NewUserRepository(db),
		emailRepo:  NewEmailCodeRepo(db),
		jwtManager: security.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration),
		cfg:        cfg,
	}
	// Используем только SMTP
	s.mailSender = NewSMTPSender(cfg)
	return s
}

func NewEmailCodeRepo(db *repo.Database) *repo.EmailCodeRepository {
	return repo.NewEmailCodeRepository(db.Pool)
}

// Register регистрация нового пользователя
func (s *AuthService) Register(ctx context.Context, req *domain.UserRegisterRequest) (*domain.AuthResponse, error) {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Нормализуем email
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.ErrInvalidField("email", "invalid email format")
	}
	if !strings.HasSuffix(email, "@"+s.cfg.AxenixEmailDomain) {
		return nil, domain.ErrInvalidField("email", fmt.Sprintf("email domain must be %s", s.cfg.AxenixEmailDomain))
	}
	// Проверка кода
	ok, err := s.emailRepo.FindValid(ctx, email, req.EmailCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInvalidInput("invalid or expired email code")
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
		Email:        email,
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

	// Помечаем код использованным и верифицируем e-mail
	if err := s.emailRepo.MarkUsed(ctx, email, req.EmailCode); err != nil {
		return nil, err
	}
	now2 := time.Now()
	user.EmailVerifiedAt = &now2

	// Генерация токена
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Очистка пароля из ответа
	user.PasswordHash = ""
	// Вычисляем isAdmin
	user.IsAdmin = s.isAdminEmail(user.Email)

	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// RequestEmailCode обрабатывает запрос одноразового кода подтверждения e-mail
func (s *AuthService) RequestEmailCode(ctx context.Context, email string, requestedIP string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return domain.ErrInvalidField("email", "email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domain.ErrInvalidField("email", "invalid email format")
	}
	if !strings.HasSuffix(email, "@"+s.cfg.AxenixEmailDomain) {
		return domain.ErrInvalidField("email", fmt.Sprintf("email domain must be %s", s.cfg.AxenixEmailDomain))
	}

	// Rate limit: не чаще 1/мин по адресу
	last, err := s.emailRepo.LastRequestAt(ctx, email)
	if err != nil {
		return err
	}
	if time.Since(last) < time.Minute {
		return domain.ErrInvalidInput("too many requests; please try again later")
	}

	// Генерация 6-значного кода
	code := generateNumericCode(6)
	ttl := time.Duration(s.cfg.EmailCodesTTLMinutes) * time.Minute
	if err := s.emailRepo.Create(ctx, email, code, ttl, requestedIP); err != nil {
		return err
	}

	// Отправка письма через mail service
	if err := s.mailSender.SendVerificationCode(ctx, email, code); err != nil {
		// Логируем подробности ошибки для диагностики
		telemetry.LoggerFromContext(ctx).Error("send verification email failed", "error", err)
		return domain.ErrInvalidInput("failed to send e-mail with code")
	}
	return nil
}

func generateNumericCode(n int) string {
	// Псевдо простой генератор; заменить на криптостойкий при необходимости
	now := time.Now().UnixNano()
	s := fmt.Sprintf("%06d", int(now%1000000))
	if len(s) > n {
		return s[len(s)-n:]
	}
	if len(s) < n {
		return strings.Repeat("0", n-len(s)) + s
	}
	return s
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
	// Вычисляем isAdmin
	user.IsAdmin = s.isAdminEmail(user.Email)

	return &domain.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// LoginByEmailCode вход или авто-регистрация по e-mail коду
func (s *AuthService) LoginByEmailCode(ctx context.Context, req *domain.EmailCodeLoginRequest) (*domain.AuthResponse, error) {
	// Валидация запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.ErrInvalidField("email", "invalid email format")
	}
	if !strings.HasSuffix(email, "@"+s.cfg.AxenixEmailDomain) {
		return nil, domain.ErrInvalidField("email", fmt.Sprintf("email domain must be %s", s.cfg.AxenixEmailDomain))
	}

	// Проверка кода
	ok, err := s.emailRepo.FindValid(ctx, email, req.EmailCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInvalidInput("invalid or expired email code")
	}

	// Пытаемся найти пользователя по e-mail
	var user *domain.User
	user, err = s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Создаём нового пользователя
			now := time.Now()
			displayName := parseDisplayNameFromEmail(email)
			login := generateLoginFromEmail(email)
			passwordHash, err := security.HashPassword(generateRandomSecret())
			if err != nil {
				return nil, errors.New("failed to hash password")
			}
			user = &domain.User{
				Login:           login,
				DisplayName:     displayName,
				Email:           email,
				PasswordHash:    passwordHash,
				EmailVerifiedAt: &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := s.userRepo.Create(ctx, user); err != nil {
				// Если конфликт логина, попробуем сгенерировать ещё несколько вариантов
				if errors.Is(err, domain.ErrConflict) {
					for i := 0; i < 1000; i++ {
						user.Login = generateLoginFromEmailWithSuffix(email, i)
						if err2 := s.userRepo.Create(ctx, user); err2 == nil {
							break
						} else if !errors.Is(err2, domain.ErrConflict) {
							return nil, err2
						}
					}
				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	// Помечаем код использованным (не критично, но желательно)
	_ = s.emailRepo.MarkUsed(ctx, email, req.EmailCode)

	// Генерируем токен на срок из конфигурации (обновим до 7 дней через cfg)
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Login, user.DisplayName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	user.PasswordHash = ""
	user.IsAdmin = s.isAdminEmail(user.Email)

	return &domain.AuthResponse{User: user, Token: token, ExpiresAt: expiresAt}, nil
}

// parseDisplayNameFromEmail извлекает имя и фамилию из local-part e-mail
func parseDisplayNameFromEmail(email string) string {
	local := strings.SplitN(email, "@", 2)[0]
	parts := strings.Split(local, ".")
	if len(parts) == 0 {
		return capitalize(local)
	}
	if len(parts) == 1 {
		return capitalize(parts[0])
	}
	// 2 части: имя + фамилия; 3+ частей: имя + последняя фамилия
	first := capitalize(parts[0])
	last := capitalize(parts[len(parts)-1])
	if last == "" {
		return first
	}
	return first + " " + last
}

func capitalize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	return strings.ToUpper(lower[:1]) + lower[1:]
}

// generateLoginFromEmail генерирует базовый логин из local-part e-mail
func generateLoginFromEmail(email string) string {
	return sanitizeLogin(strings.SplitN(email, "@", 2)[0])
}

// generateLoginFromEmailWithSuffix генерирует логин с числовым суффиксом
func generateLoginFromEmailWithSuffix(email string, n int) string {
	base := sanitizeLogin(strings.SplitN(email, "@", 2)[0])
	if n <= 0 {
		return base
	}
	return fmt.Sprintf("%s-%03d", base, n)
}

// sanitizeLogin оставляет только [a-z0-9._-] и обрезает до 32 символов
func sanitizeLogin(local string) string {
	local = strings.ToLower(local)
	b := make([]rune, 0, len(local))
	for _, r := range local {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b = append(b, r)
		}
	}
	s := string(b)
	if len(s) > 32 {
		s = s[:32]
	}
	if s == "" {
		s = "user"
	}
	return s
}

// generateRandomSecret генерирует случайную строку для захешированного пароля
func generateRandomSecret() string {
	// Используем текущее время как простой источник; для продакшена можно заменить на криптостойкий
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// GetUserByID получение пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Очистка пароля
	user.PasswordHash = ""
	// Вычисляем isAdmin
	user.IsAdmin = s.isAdminEmail(user.Email)
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

// isAdminEmail проверяет, входит ли e-mail пользователя в список админов
func (s *AuthService) isAdminEmail(email string) bool {
	e := strings.ToLower(strings.TrimSpace(email))
	for _, admin := range s.cfg.AdminEmails {
		if e == admin {
			return true
		}
	}
	return false
}
