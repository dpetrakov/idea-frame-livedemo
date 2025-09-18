package http

import (
	"log/slog"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/config"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/http/handlers"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter создает новый HTTP роутер с настроенными маршрутами
func NewRouter(cfg *config.Config, logger *slog.Logger, authService *service.AuthService, pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))
	r.Use(middleware.Timeout(15 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // В продакшене должен быть настроен точнее
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Создание handlers
	authHandler := handlers.NewAuthHandler(authService, logger, pool)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health endpoint (без аутентификации)
		r.Get("/health", authHandler.Health)

		// Аутентификация (без аутентификации)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Защищённые маршруты (требуют JWT)
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(authService))

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", authHandler.GetCurrentUser)
				r.Get("/", authHandler.GetUsers) // Для назначения ответственных
			})
		})
	})

	// Алиас для health endpoint для совместимости (согласно TK-001 спецификации)
	r.Get("/api/health", authHandler.Health)

	return r
}
