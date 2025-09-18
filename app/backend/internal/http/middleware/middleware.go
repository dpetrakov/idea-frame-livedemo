package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
	
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/service"
	"github.com/ideaframe/backend/internal/telemetry"
)

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// UserIDKey ключ для ID пользователя в контексте
	UserIDKey ContextKey = "userID"
	// RequestIDKey ключ для ID запроса в контексте
	RequestIDKey ContextKey = "requestID"
)

// ErrorResponse структура ответа с ошибкой
type ErrorResponse struct {
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Details       interface{} `json:"details,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// RespondWithError отправляет JSON ответ с ошибкой
func RespondWithError(w http.ResponseWriter, r *http.Request, code int, message string, errCode string) {
	response := ErrorResponse{
		Code:    errCode,
		Message: message,
	}
	
	// Добавляем correlation ID если есть
	if reqID := r.Context().Value(RequestIDKey); reqID != nil {
		response.CorrelationID = reqID.(string)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithErrorDetails отправляет JSON ошибку с details
func RespondWithErrorDetails(w http.ResponseWriter, r *http.Request, code int, message string, errCode string, details interface{}) {
	response := ErrorResponse{
		Code:    errCode,
		Message: message,
		Details: details,
	}
	if reqID := r.Context().Value(RequestIDKey); reqID != nil {
		response.CorrelationID = reqID.(string)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RequestID добавляет уникальный ID запроса
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "req_" + uuid.New().String()
		}
		
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger логирует HTTP запросы
func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем логгер в контекст
			ctx := context.WithValue(r.Context(), telemetry.LoggerKey, logger)
			r = r.WithContext(ctx)
			
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			defer func() {
				reqID := r.Context().Value(RequestIDKey)
				logger.Info("request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration_ms", time.Since(start).Milliseconds(),
					"request_id", reqID,
				)
			}()
			
			next.ServeHTTP(ww, r)
		})
	}
}

// Auth проверяет JWT токен и добавляет userID в контекст
func Auth(authService *service.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondWithError(w, r, http.StatusUnauthorized, "Authorization header required", "UNAUTHORIZED")
				return
			}
			
			// Проверяем формат: Bearer <token>
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				RespondWithError(w, r, http.StatusUnauthorized, "Invalid authorization header format", "UNAUTHORIZED")
				return
			}
			
			token := parts[1]
			
			// Валидируем токен
			claims, err := authService.ValidateToken(token)
			if err != nil {
				RespondWithError(w, r, http.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
				return
			}
			
			// Добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS настройка CORS заголовков
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любых источников для демо
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext получает ID пользователя из контекста
func UserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

// RequestIDFromContext получает ID запроса из контекста
func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// Recoverer восстанавливается после паники
func Recoverer(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID := r.Context().Value(RequestIDKey)
					logger.Error("panic recovered",
						"error", err,
						"request_id", reqID,
						"path", r.URL.Path,
					)
					RespondWithError(w, r, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}