package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/http/middleware"
	"github.com/ideaframe/backend/internal/service"
)

// AuthHandler обработчики аутентификации
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler создаёт новый обработчик аутентификации
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обработчик регистрации POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.UserRegisterRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, r, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}
	
	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		var validationErr domain.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Message == "user with this login already exists" {
				middleware.RespondWithError(w, r, http.StatusConflict, validationErr.Message, "CONFLICT")
			} else {
				middleware.RespondWithError(w, r, http.StatusBadRequest, validationErr.Message, "VALIDATION_ERROR")
			}
			return
		}
		
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to register user", "INTERNAL_ERROR")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login обработчик входа POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.UserLoginRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, r, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}
	
	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			middleware.RespondWithError(w, r, http.StatusUnauthorized, "Invalid credentials", "INVALID_CREDENTIALS")
			return
		}
		
		var validationErr domain.ValidationError
		if errors.As(err, &validationErr) {
			middleware.RespondWithError(w, r, http.StatusBadRequest, validationErr.Message, "VALIDATION_ERROR")
			return
		}
		
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to login", "INTERNAL_ERROR")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetCurrentUser обработчик получения текущего пользователя GET /users/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (добавлен middleware Auth)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		middleware.RespondWithError(w, r, http.StatusUnauthorized, "User not authenticated", "UNAUTHORIZED")
		return
	}
	
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			middleware.RespondWithError(w, r, http.StatusNotFound, "User not found", "NOT_FOUND")
			return
		}
		
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get user", "INTERNAL_ERROR")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// ListUsers обработчик списка пользователей GET /users
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.authService.ListUsers(r.Context())
	if err != nil {
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to list users", "INTERNAL_ERROR")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}