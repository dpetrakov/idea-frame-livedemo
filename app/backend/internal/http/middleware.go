package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
	"github.com/go-chi/chi/v5/middleware"
)

// ErrorResponse представляет формат ошибки согласно OpenAPI
type ErrorResponse struct {
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Details       interface{} `json:"details,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// AuthMiddleware создает middleware для проверки JWT токенов
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil, getRequestID(r))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Некорректный формат токена", nil, getRequestID(r))
				return
			}

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Недействительный токен", nil, getRequestID(r))
				return
			}

			// Добавляем ID пользователя в контекст
			ctx := context.WithValue(r.Context(), security.UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware создает middleware для логирования запросов
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return middleware.RequestLogger(&structuredLogger{logger})
}

// structuredLogger реализует middleware.LogFormatter для slog
type structuredLogger struct {
	Logger *slog.Logger
}

func (l *structuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &structuredLogEntry{
		Logger:    l.Logger,
		request:   r,
		requestID: getRequestID(r),
	}
}

type structuredLogEntry struct {
	Logger    *slog.Logger
	request   *http.Request
	requestID string
}

func (l *structuredLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger.Info("http_request",
		"method", l.request.Method,
		"path", l.request.URL.Path,
		"status", status,
		"bytes", bytes,
		"elapsed_ms", elapsed.Milliseconds(),
		"request_id", l.requestID,
	)
}

func (l *structuredLogEntry) Panic(v interface{}, stack []byte) {
	l.Logger.Error("http_panic",
		"panic", fmt.Sprintf("%+v", v),
		"stack", string(stack),
		"request_id", l.requestID,
	)
}

// RecoveryMiddleware обрабатывает панки и возвращает JSON ошибку
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", fmt.Sprintf("%+v", err),
						"request_id", getRequestID(r),
						"path", r.URL.Path,
					)

					writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
						"Внутренняя ошибка сервера", nil, getRequestID(r))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// writeError отправляет JSON ошибку клиенту
func writeError(w http.ResponseWriter, status int, code, message string, details interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(status)

	response := ErrorResponse{
		Code:          code,
		Message:       message,
		Details:       details,
		CorrelationID: requestID,
	}

	json.NewEncoder(w).Encode(response)
}

// getRequestID извлекает request ID из контекста
func getRequestID(r *http.Request) string {
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		return reqID
	}
	return "unknown"
}
