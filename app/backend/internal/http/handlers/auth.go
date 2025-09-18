package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/repo"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
	pool        *pgxpool.Pool
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(authService *service.AuthService, logger *slog.Logger, pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		pool:        pool,
	}
}

// HealthResponse представляет ответ health check согласно OpenAPI
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

// Health проверяет состояние сервиса
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := "ok"
	dbStatus := "connected"

	// Проверка подключения к БД
	if err := repo.HealthCheck(ctx, h.pool); err != nil {
		h.logger.Error("database health check failed", "error", err)
		dbStatus = "disconnected"
		status = "error"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	response := HealthResponse{
		Status:    status,
		Database:  dbStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	json.NewEncoder(w).Encode(response)
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный JSON", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	authResponse, err := h.authService.Register(ctx, &req)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	h.logger.Info("user registered successfully",
		"user_id", authResponse.User.ID,
		"login", authResponse.User.Login,
		"request_id", middleware.GetReqID(r.Context()))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Некорректный JSON", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	authResponse, err := h.authService.Login(ctx, &req)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	h.logger.Info("user logged in successfully",
		"user_id", authResponse.User.ID,
		"login", authResponse.User.Login,
		"request_id", middleware.GetReqID(r.Context()))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponse)
}

// GetCurrentUser возвращает информацию о текущем пользователе
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := security.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Пользователь не найден в контексте", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.authService.GetCurrentUser(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			h.writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Пользователь не найден", nil)
			return
		}
		h.logger.Error("failed to get current user", "error", err, "user_id", userID)
		h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// GetUsers возвращает список всех пользователей
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users, err := h.authService.GetAllUsers(ctx)
	if err != nil {
		h.logger.Error("failed to get users", "error", err)
		h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// handleServiceError обрабатывает ошибки сервисного слоя
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr service.ValidationError
	if errors.As(err, &validationErr) {
		details := map[string]string{
			"field":         validationErr.Field,
			"rejectedValue": "",
		}
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", validationErr.Message, details)
		return
	}

	if errors.Is(err, domain.ErrUserAlreadyExists) {
		h.writeError(w, r, http.StatusConflict, "CONFLICT", "Пользователь уже существует", nil)
		return
	}

	if errors.Is(err, domain.ErrInvalidCredentials) {
		h.writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Неверные учетные данные", nil)
		return
	}

	h.logger.Error("service error", "error", err)
	h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
}

// ErrorResponse представляет формат ошибки согласно OpenAPI
type ErrorResponse struct {
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Details       interface{} `json:"details,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// writeError отправляет JSON ошибку клиенту
func (h *AuthHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(status)

	response := ErrorResponse{
		Code:          code,
		Message:       message,
		Details:       details,
		CorrelationID: middleware.GetReqID(r.Context()),
	}

	json.NewEncoder(w).Encode(response)
}
