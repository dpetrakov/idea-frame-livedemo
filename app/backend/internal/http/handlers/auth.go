package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
)

type contextKey string

const (
	RequestIDKey contextKey = "requestId"
	UserIDKey    contextKey = "userId"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register обрабатывает POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req domain.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	authResp, err := h.authService.Register(ctx, &req)
	if err != nil {
		h.logger.Warn("registration failed",
			"login", req.Login,
			"error", err.Error(),
			"request_id", r.Context().Value(RequestIDKey),
		)

		if errors.Is(err, domain.ErrUserAlreadyExists) {
			h.writeErrorResponse(w, r, http.StatusConflict, "USER_ALREADY_EXISTS", "User with this login already exists")
			return
		}

		// Ошибки валидации
		if isValidationError(err) {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}

		h.writeErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Registration failed")
		return
	}

	h.logger.Info("user registered successfully",
		"user_id", authResp.User.ID,
		"login", authResp.User.Login,
		"request_id", r.Context().Value(RequestIDKey),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResp)
}

// Login обрабатывает POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req domain.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	authResp, err := h.authService.Login(ctx, &req)
	if err != nil {
		h.logger.Warn("login failed",
			"login", req.Login,
			"error", err.Error(),
			"request_id", r.Context().Value(RequestIDKey),
		)

		if errors.Is(err, domain.ErrInvalidPassword) {
			h.writeErrorResponse(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid login or password")
			return
		}

		// Ошибки валидации
		if isValidationError(err) {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}

		h.writeErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed")
		return
	}

	h.logger.Info("user logged in successfully",
		"user_id", authResp.User.ID,
		"login", authResp.User.Login,
		"request_id", r.Context().Value(RequestIDKey),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResp)
}

// GetCurrentUser обрабатывает GET /users/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		h.writeErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in context")
		return
	}

	user, err := h.authService.GetCurrentUser(ctx, userID)
	if err != nil {
		h.logger.Warn("get current user failed",
			"user_id", userID,
			"error", err.Error(),
			"request_id", r.Context().Value(RequestIDKey),
		)

		if errors.Is(err, domain.ErrUserNotFound) {
			h.writeErrorResponse(w, r, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		h.writeErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// writeErrorResponse записывает ошибку в ответ
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	requestID, _ := r.Context().Value(RequestIDKey).(string)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var correlationID uuid.UUID
	if requestID != "" {
		if parsed, err := uuid.Parse(requestID); err == nil {
			correlationID = parsed
		} else {
			correlationID = uuid.New()
		}
	} else {
		correlationID = uuid.New()
	}

	errorResp := domain.NewErrorResponse(code, message, correlationID)
	json.NewEncoder(w).Encode(errorResp)
}

// isValidationError проверяет, является ли ошибка ошибкой валидации
func isValidationError(err error) bool {
	// Простая проверка на основе текста ошибки
	// В реальном проекте лучше создать специальные типы ошибок
	errMsg := err.Error()
	return errMsg != "" && 
		   (errMsg[:10] == "login must" ||
			errMsg[:12] == "display name" ||
			errMsg[:8] == "password" ||
			errMsg[:9] == "passwords")
}