package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/config"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/http/handlers"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool, jwtService *security.JWTService, authService *service.AuthService) http.Handler {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(RequestIDMiddleware)
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))
	r.Use(CORSMiddleware)
	r.Use(middleware.Timeout(15 * time.Second))

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	healthHandler := handlers.NewHealthHandler(pool)

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.Health)

		// Authentication
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(jwtService))
			r.Get("/users/me", authHandler.GetCurrentUser)
		})
	})

	// Alias для health check согласно требованиям
	r.Get("/api/health", healthHandler.Health)

	return r
}