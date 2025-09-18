package http

import (
	"log/slog"
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ideaframe/backend/internal/config"
	"github.com/ideaframe/backend/internal/http/handlers"
	"github.com/ideaframe/backend/internal/http/middleware"
	"github.com/ideaframe/backend/internal/repo"
	"github.com/ideaframe/backend/internal/service"
)

// NewRouter создаёт и настраивает HTTP роутер
func NewRouter(
	cfg *config.Config,
	logger *slog.Logger,
	db *repo.Database,
) http.Handler {
	r := chi.NewRouter()
	
	// Создаём сервисы
	authService := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiration)
	
	// Создаём handlers
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(db)
	
	// Глобальные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS)
	
	// API версия 1
	r.Route("/api/v1", func(r chi.Router) {
		// Health check (публичный)
		r.Get("/health", healthHandler.Health)
		
		// Аутентификация (публичные)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
		
		// Защищённые маршруты
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			
			// Пользователи
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", authHandler.GetCurrentUser)
				r.Get("/", authHandler.ListUsers)
			})
		})
	})
	
	// Health check для /api/health (алиас для совместимости)
	r.Get("/api/health", healthHandler.Health)
	
	return r
}