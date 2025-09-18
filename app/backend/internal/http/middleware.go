package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
)

type contextKey string

const (
	RequestIDKey contextKey = "requestId"
	UserIDKey    contextKey = "userId"
)

// RequestID добавляет уникальный ID запроса в контекст и заголовок
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware логирует HTTP запросы
func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := r.Context().Value(RequestIDKey).(string)

			// Логируем начало запроса
			logger.Info("request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"request_id", requestID,
			)

			defer func() {
				// Логируем завершение запроса
				duration := time.Since(start)
				logger.Info("request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", duration.Milliseconds(),
					"request_id", requestID,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// RecoveryMiddleware обрабатывает панику
func RecoveryMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID, _ := r.Context().Value(RequestIDKey).(string)
					
					logger.Error("panic recovered",
						"error", rec,
						"method", r.Method,
						"path", r.URL.Path,
						"request_id", requestID,
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					errorResp := domain.NewErrorResponse(
						"INTERNAL_SERVER_ERROR",
						"Internal server error",
						uuid.MustParse(requestID),
					)
					json.NewEncoder(w).Encode(errorResp)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(jwtService *security.JWTService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				writeErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
				return
			}

			claims, err := jwtService.ValidateToken(bearerToken[1])
			if err != nil {
				writeErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
				return
			}

			// Добавляем ID пользователя в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse помогает записать ошибку в ответ
func writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
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